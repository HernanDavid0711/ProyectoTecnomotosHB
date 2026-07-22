package httpserver

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"tecnomotos/internal/auth"
	"tecnomotos/internal/clientes"
	"tecnomotos/internal/config"
	"tecnomotos/internal/motos"
	"tecnomotos/internal/ordenes"
	"tecnomotos/internal/sesion"
	"tecnomotos/internal/shared"
	"tecnomotos/internal/usuarios"
)

type Server struct {
	httpServer *http.Server
	db         *pgxpool.Pool
	cfg        *config.Config
}

func New(cfg *config.Config, db *pgxpool.Pool) *Server {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), corsMiddleware())

	s := &Server{
		db:  db,
		cfg: cfg,
	}

	router.GET("/health", ginHTTP(s.healthHandler))
	router.GET("/", ginHTTP(s.homeHandler))

	authMiddleware := auth.Authenticate(cfg.JWTSecret, cfg.JWTIssuer)
	adminMiddleware := auth.RequireRole(usuarios.RolAdministrador)
	creadorOrdenesMiddleware := auth.RequireRole(usuarios.RolAdministrador, usuarios.RolCreadorOrdenes)
	empleadoMiddleware := auth.RequireRole(usuarios.RolEmpleado)

	usuarioRepo := usuarios.NewRepository(db)
	usuarioService := usuarios.NewService(usuarioRepo)
	if err := usuarioService.EnsureDefaultRoles(context.Background()); err != nil {
		log.Printf("error asegurando roles base: %v", err)
	}
	usuarioHandler := usuarios.NewHandler(usuarioService)
	router.GET("/api/admin/roles", ginHTTP(usuarioHandler.ListRoles, authMiddleware, adminMiddleware))
	router.GET("/api/admin/usuarios", ginHTTP(usuarioHandler.List, authMiddleware, adminMiddleware))
	router.POST("/api/admin/usuarios", ginHTTP(usuarioHandler.Create, authMiddleware, adminMiddleware))
	router.GET("/api/admin/usuarios/:id", ginHTTP(usuarioHandler.GetByID, authMiddleware, adminMiddleware))
	router.PUT("/api/admin/usuarios/:id", ginHTTP(usuarioHandler.Update, authMiddleware, adminMiddleware))
	router.PATCH("/api/admin/usuarios/:id/rol", ginHTTP(usuarioHandler.AssignRole, authMiddleware, adminMiddleware))
	router.DELETE("/api/admin/usuarios/:id", ginHTTP(usuarioHandler.Deactivate, authMiddleware, adminMiddleware))

	sesionHandler := sesion.NewHandler(cfg, usuarioService)
	router.POST("/api/auth/bootstrap-admin", ginHTTP(sesionHandler.BootstrapAdmin))
	router.POST("/api/auth/login", ginHTTP(sesionHandler.Login))
	router.GET("/api/auth/me", ginHTTP(sesionHandler.Me, authMiddleware))

	clienteRepo := clientes.NewRepository(db)
	clienteService := clientes.NewService(clienteRepo)
	clienteHandler := clientes.NewHandler(clienteService)
	router.GET("/api/clientes", ginHTTP(clienteHandler.List, authMiddleware, adminMiddleware))
	router.POST("/api/clientes", ginHTTP(clienteHandler.Create, authMiddleware, adminMiddleware))
	router.GET("/api/clientes/:id", ginHTTP(clienteHandler.GetByID, authMiddleware, adminMiddleware))
	router.PUT("/api/clientes/:id", ginHTTP(clienteHandler.Update, authMiddleware, adminMiddleware))
	router.DELETE("/api/clientes/:id", ginHTTP(clienteHandler.Delete, authMiddleware, adminMiddleware))

	motoRepo := motos.NewRepository(db)
	motoService := motos.NewService(motoRepo)
	motoHandler := motos.NewHandler(motoService)
	router.GET("/api/motos", ginHTTP(motoHandler.List, authMiddleware, adminMiddleware))
	router.POST("/api/motos", ginHTTP(motoHandler.Create, authMiddleware, adminMiddleware))
	router.GET("/api/motos/autocompletar", ginHTTP(motoHandler.AutocompleteByPlaca, authMiddleware, adminMiddleware))
	router.GET("/api/motos/por-placa/:placa", ginHTTP(motoHandler.GetByPlaca, authMiddleware, adminMiddleware))
	router.GET("/api/motos/:id", ginHTTP(motoHandler.GetByID, authMiddleware, adminMiddleware))
	router.PUT("/api/motos/:id", ginHTTP(motoHandler.Update, authMiddleware, adminMiddleware))

	ordenRepo := ordenes.NewRepository(db)
	ordenService := ordenes.NewService(ordenRepo)
	ordenHandler := ordenes.NewHandler(ordenService)
	router.GET("/api/admin/formato-orden", ginHTTP(ordenHandler.GetFormato, authMiddleware, adminMiddleware))
	router.GET("/api/admin/catalogo-items", ginHTTP(ordenHandler.AdminListCatalogoItems, authMiddleware, adminMiddleware))
	router.POST("/api/admin/catalogo-items", ginHTTP(ordenHandler.AdminCreateCatalogoItem, authMiddleware, adminMiddleware))
	router.GET("/api/admin/catalogo-items/:id", ginHTTP(ordenHandler.AdminGetCatalogoItem, authMiddleware, adminMiddleware))
	router.PUT("/api/admin/catalogo-items/:id", ginHTTP(ordenHandler.AdminUpdateCatalogoItem, authMiddleware, adminMiddleware))
	router.DELETE("/api/admin/catalogo-items/:id", ginHTTP(ordenHandler.AdminDeleteCatalogoItem, authMiddleware, adminMiddleware))
	router.GET("/api/admin/ordenes", ginHTTP(ordenHandler.AdminList, authMiddleware, adminMiddleware))
	router.POST("/api/admin/ordenes", ginHTTP(ordenHandler.AdminCreate, authMiddleware, adminMiddleware))
	router.GET("/api/admin/ordenes/:id", ginHTTP(ordenHandler.AdminGetByID, authMiddleware, adminMiddleware))
	router.PATCH("/api/admin/ordenes/:id", ginHTTP(ordenHandler.AdminUpdate, authMiddleware, adminMiddleware))
	router.PATCH("/api/admin/ordenes/:id/asignacion", ginHTTP(ordenHandler.AdminAssign, authMiddleware, adminMiddleware))
	router.GET("/api/creador-ordenes/clientes", ginHTTP(ordenHandler.BuscarClientesConMotos, authMiddleware, creadorOrdenesMiddleware))
	router.GET("/api/creador-ordenes/formato-orden", ginHTTP(ordenHandler.GetFormato, authMiddleware, creadorOrdenesMiddleware))
	router.GET("/api/creador-ordenes/catalogo-items", ginHTTP(ordenHandler.ListCatalogoItems, authMiddleware, creadorOrdenesMiddleware))
	router.GET("/api/creador-ordenes/empleados", ginHTTP(ordenHandler.ListEmpleadosAsignables, authMiddleware, creadorOrdenesMiddleware))
	router.POST("/api/creador-ordenes/ordenes", ginHTTP(ordenHandler.CreadorCreateWithItems, authMiddleware, creadorOrdenesMiddleware))
	router.GET("/api/empleado/ordenes", ginHTTP(ordenHandler.EmpleadoList, authMiddleware, empleadoMiddleware))
	router.GET("/api/empleado/ordenes/:id", ginHTTP(ordenHandler.EmpleadoGetByID, authMiddleware, empleadoMiddleware))
	router.PATCH("/api/empleado/ordenes/:id/progreso", ginHTTP(ordenHandler.EmpleadoUpdateProgress, authMiddleware, empleadoMiddleware))

	s.httpServer = &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return s
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func ginHTTP(handler http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, param := range c.Params {
			c.Request.SetPathValue(param.Key, param.Value)
		}

		var httpHandler http.Handler = handler
		for i := len(middlewares) - 1; i >= 0; i-- {
			httpHandler = middlewares[i](httpHandler)
		}

		httpHandler.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "API Tecnomotos funcionando",
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		shared.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status":   "error",
			"database": "down",
			"error":    err.Error(),
		})
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"database": "up",
	})
}
