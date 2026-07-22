const API_BASE_URL = window.TECNOMOTOS_API_BASE_URL || "http://localhost:8081";

const loginPage = document.querySelector("#login-page");
const adminApp = document.querySelector("#admin-app");
const loginForm = document.querySelector("#login-form");
const usuarioInput = document.querySelector("#usuario");
const contrasenaInput = document.querySelector("#contrasena");
const togglePasswordButton = document.querySelector("#toggle-password");
const submitButton = document.querySelector("#submit-button");
const formError = document.querySelector("#form-error");
const adminUserName = document.querySelector("#admin-user-name");
const adminUserRole = document.querySelector("#admin-user-role");
const sidebarRoleLabel = document.querySelector("#sidebar-role-label");
const sectionTitle = document.querySelector("#section-title");
const sectionAction = document.querySelector("#section-action");
const adminContent = document.querySelector("#admin-content");
const navItems = document.querySelectorAll(".nav-item");
const logoutButton = document.querySelector("#logout-button");

const storageKeys = {
  token: "tecnomotos_token",
  user: "tecnomotos_user",
  expiresAt: "tecnomotos_expires_at"
};

const ordenState = {
  clientes: [],
  selectedCliente: null,
  selectedMoto: null,
  empleados: [],
  selectedEmpleado: "",
  catalogoItems: [],
  selectedItems: [],
  formato: null,
  formatoLoaded: false
};

const adminOrdenState = {
  ordenes: [],
  selectedOrden: null,
  empleados: [],
  ordenesLoaded: false,
  empleadosLoaded: false,
  loading: false,
  formato: null,
  formatoLoaded: false
};

const defaultFormatoCampos = [
  { id: "kilometraje_ingreso", label: "Kilometraje", tipo: "numero", requerido: true, sistema: true },
  { id: "descripcion_falla", label: "Descripcion de falla", tipo: "area", requerido: true, sistema: true },
  { id: "observaciones", label: "Observaciones", tipo: "area", requerido: false, sistema: true }
];

function getSession() {
  const token = localStorage.getItem(storageKeys.token);
  const user = localStorage.getItem(storageKeys.user);

  if (!token || !user) return null;

  try {
    return {
      token,
      user: JSON.parse(user),
      expiresAt: localStorage.getItem(storageKeys.expiresAt)
    };
  } catch {
    clearSession();
    return null;
  }
}

function saveSession(payload) {
  localStorage.setItem(storageKeys.token, payload.token);
  localStorage.setItem(storageKeys.user, JSON.stringify(payload.data));
  localStorage.setItem(storageKeys.expiresAt, payload.expires_at);
}

function clearSession() {
  localStorage.removeItem(storageKeys.token);
  localStorage.removeItem(storageKeys.user);
  localStorage.removeItem(storageKeys.expiresAt);
}

function authHeaders() {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${localStorage.getItem(storageKeys.token)}`
  };
}

async function apiGet(path) {
  const response = await fetch(`${API_BASE_URL}${path}`, { headers: authHeaders() });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.message || "No se pudo consultar la informacion");
  return payload;
}

async function apiPost(path, body) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(body)
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.message || "No se pudo guardar la informacion");
  return payload;
}

async function apiPatch(path, body) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: "PATCH",
    headers: authHeaders(),
    body: JSON.stringify(body)
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.message || "No se pudo actualizar la informacion");
  return payload;
}

async function login(usuario, contrasena) {
  const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ usuario, contrasena })
  });

  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.message || "No se pudo iniciar sesion");
  return payload;
}

function renderSession(session) {
  if (!session) {
    loginPage.hidden = false;
    adminApp.hidden = true;
    return;
  }

  const nombres = [session.user?.nombres, session.user?.apellidos].filter(Boolean).join(" ");
  adminUserName.textContent = nombres || session.user?.usuario || session.user?.correo || "Usuario";
  adminUserRole.textContent = session.user?.rol || "sin rol";

  loginPage.hidden = true;
  adminApp.hidden = false;

  const isOrderCreator = session.user?.rol === "creador_ordenes";
  adminApp.classList.toggle("creator-profile", isOrderCreator);
  sidebarRoleLabel.textContent = isOrderCreator ? "Perfil creador de ordenes" : "Panel administrador";

  if (isOrderCreator) {
    setActiveSection("ordenes");
  } else if (getActiveSection() === "ordenes") {
    setActiveSection("usuarios");
  }

  renderAdminSection(getActiveSection());
}

function setError(message) {
  formError.textContent = message;
  formError.hidden = !message;
}

function setLoading(loading) {
  submitButton.disabled = loading;
  submitButton.textContent = loading ? "Validando..." : "Entrar";
}

togglePasswordButton.addEventListener("click", () => {
  const showing = contrasenaInput.type === "text";
  contrasenaInput.type = showing ? "password" : "text";
  togglePasswordButton.textContent = showing ? "Ver" : "Ocultar";
  togglePasswordButton.setAttribute("aria-label", showing ? "Mostrar contrasena" : "Ocultar contrasena");
  togglePasswordButton.title = showing ? "Mostrar contrasena" : "Ocultar contrasena";
});

loginForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  const usuario = usuarioInput.value.trim();
  const contrasena = contrasenaInput.value.trim();
  if (!usuario || !contrasena) {
    setError("Ingresa usuario y contrasena.");
    return;
  }

  setLoading(true);
  setError("");

  try {
    const payload = await login(usuario, contrasena);
    saveSession(payload);
    renderSession({ token: payload.token, user: payload.data, expiresAt: payload.expires_at });
  } catch (error) {
    setError(error.message);
  } finally {
    setLoading(false);
  }
});

logoutButton.addEventListener("click", () => {
  clearSession();
  loginForm.reset();
  renderSession(null);
});

navItems.forEach((item) => {
  item.addEventListener("click", () => {
    setActiveSection(item.dataset.section);
    renderAdminSection(item.dataset.section);
  });
});

function setActiveSection(section) {
  navItems.forEach((item) => item.classList.toggle("active", item.dataset.section === section));
}

function getActiveSection() {
  return document.querySelector(".nav-item.active")?.dataset.section || "usuarios";
}

function renderAdminSection(section) {
  const sections = {
    usuarios: {
      title: "Usuarios",
      action: "Nuevo usuario",
      eyebrow: "Control de acceso",
      heading: "Gestion de usuarios del sistema",
      body: "Aqui construiremos el listado, creacion, edicion, desactivacion y asignacion de roles para cada usuario.",
      stats: [["Rol principal", "Administrador"], ["Seguridad", "JWT activo"], ["Estado", "Preparado"]],
      table: ["Usuario", "Nombre", "Rol", "Estado"]
    },
    ordenes: {
      title: "Ordenes de trabajo",
      action: getSession()?.user?.rol === "creador_ordenes" ? "Crear orden" : "Vista formato",
      custom: "ordenes"
    },
    catalogos: {
      title: "Catalogos",
      action: "Nuevo catalogo",
      eyebrow: "Datos base",
      heading: "Catalogos operativos del taller",
      body: "Este espacio quedara para marcas, modelos, tipos de servicio, estados, metodos de pago y valores reutilizables.",
      stats: [["Marcas", "Pendiente"], ["Servicios", "Pendiente"], ["Repuestos", "Pendiente"]],
      table: ["Catalogo", "Descripcion", "Items", "Estado"]
    },
    empleados: {
      title: "Empleados",
      action: "Nuevo empleado",
      eyebrow: "Equipo de trabajo",
      heading: "Administracion de empleados",
      body: "Aqui vamos a separar los usuarios con rol empleado, revisar sus ordenes asignadas y su carga de trabajo.",
      stats: [["Rol", "Empleado"], ["Ordenes asignadas", "Pendiente"], ["Productividad", "Pendiente"]],
      table: ["Empleado", "Correo", "Ordenes", "Disponibilidad"]
    },
    clientes: {
      title: "Clientes",
      action: "Nuevo cliente",
      eyebrow: "Relacion con clientes",
      heading: "Clientes registrados",
      body: "Esta seccion conectara con el modulo de clientes para consultar, crear y actualizar informacion de contacto.",
      stats: [["Activos", "Pendiente"], ["Documentos", "Validado"], ["Contacto", "Centralizado"]],
      table: ["Cliente", "Documento", "Telefono", "Estado"]
    },
    historial: {
      title: "Historial de motocicletas",
      action: "Buscar moto",
      eyebrow: "Trazabilidad",
      heading: "Historial de motocicletas",
      body: "Aqui construiremos la consulta por placa, propietario, ordenes realizadas, kilometraje y trabajos previos.",
      stats: [["Busqueda", "Por placa"], ["Ordenes", "Historicas"], ["Propietarios", "Trazables"]],
      table: ["Placa", "Cliente", "Ultima orden", "Kilometraje"]
    }
  };

  const current = sections[section] || sections.usuarios;
  sectionTitle.textContent = current.title;
  sectionAction.textContent = current.action;

  if (current.custom === "ordenes") {
    const session = getSession();
    if (session?.user?.rol === "creador_ordenes") {
      renderOrdenesSection();
    } else {
      renderAdminOrdenesSection();
    }
    return;
  }

  adminContent.innerHTML = `
    <div class="content-intro">
      <div>
        <p class="content-eyebrow">${current.eyebrow}</p>
        <h2>${current.heading}</h2>
        <p>${current.body}</p>
      </div>
    </div>
    <div class="metric-grid">
      ${current.stats.map(([label, value]) => `<article class="metric-card"><span>${label}</span><strong>${value}</strong></article>`).join("")}
    </div>
    <section class="work-panel">
      <div class="panel-toolbar">
        <div>
          <h3>${current.title}</h3>
          <p>Vista inicial lista para conectar con la API.</p>
        </div>
        <input class="search-input" type="search" placeholder="Buscar" />
      </div>
      <div class="data-table" role="table" aria-label="${current.title}">
        <div class="table-row table-head" role="row">
          ${current.table.map((column) => `<span role="columnheader">${column}</span>`).join("")}
        </div>
        <div class="empty-state">
          <strong>Modulo en construccion</strong>
          <span>Seguimos con esta pestana en el proximo paso.</span>
        </div>
      </div>
    </section>
  `;
}

function renderAdminOrdenesSection() {
  const campos = adminOrdenState.formato?.campos || defaultFormatoCampos;
  adminContent.innerHTML = `
    <div class="content-intro">
      <div>
        <p class="content-eyebrow">Vista del formato</p>
        <h2>Orden de trabajo para tablet</h2>
        <p>Este es el formato que vera el rol creador de ordenes de trabajo al registrar una moto en taller.</p>
      </div>
    </div>

    <section class="order-document admin-format-preview">
      <header class="order-document-header">
        <div>
          <span>Formato operativo</span>
          <strong>Orden de trabajo</strong>
        </div>
        <div>
          <span>Rol que lo diligencia</span>
          <strong>Creador</strong>
        </div>
      </header>

      <div class="tablet-preview-grid">
        <section class="preview-block">
          <div>
            <h3>Cliente y motocicleta</h3>
            <p>Busqueda por nombre, telefono o placa.</p>
          </div>
          <label class="field compact-field"><span>Buscar cliente</span><input class="plain-input" type="text" value="Nombre, telefono o placa" disabled /></label>
          <label class="field compact-field"><span>Motocicleta</span><select class="plain-input" disabled><option>Placa - marca - modelo</option></select></label>
        </section>

        <section class="preview-block">
          <div>
            <h3>Datos de ingreso</h3>
            <p>Campos principales de la orden.</p>
          </div>
          ${campos.map((campo) => renderDynamicField(campo, "preview", true)).join("")}
        </section>

        <section class="preview-block">
          <div>
            <h3>Asignacion e items</h3>
            <p>Items seleccionables desde catalogo de servicios.</p>
          </div>
          <label class="field compact-field"><span>Empleado responsable</span><select class="plain-input" disabled><option>Empleado asignado</option></select></label>
          <div class="selected-items">
            <div class="selected-item">
              <select class="plain-input item-select" disabled><option>Revision general - mano_obra - $0</option></select>
              <input class="mini-input" type="number" value="1" disabled />
              <input class="mini-input" type="number" value="0" disabled />
              <button class="remove-item-button" type="button" disabled>Quitar</button>
            </div>
          </div>
        </section>
      </div>
    </section>
  `;

  if (!adminOrdenState.formatoLoaded) {
    cargarFormatoAdmin();
  }
}

async function cargarFormatoAdmin() {
  try {
    const payload = await apiGet("/api/admin/formato-orden");
    adminOrdenState.formato = payload.data || { campos: defaultFormatoCampos };
  } catch {
    adminOrdenState.formato = { campos: defaultFormatoCampos };
  } finally {
    adminOrdenState.formatoLoaded = true;
    renderAdminOrdenesSection();
  }
}

function bindAdminOrdenesEvents() {
  document.querySelector("#admin-buscar-ordenes-btn")?.addEventListener("click", cargarAdminOrdenes);
  document.querySelector("#admin-guardar-orden-btn")?.addEventListener("click", guardarAdminOrden);
  document.querySelector("#admin-recargar-orden-btn")?.addEventListener("click", () => renderAdminOrdenesSection());
}

async function cargarAdminOrdenes() {
  const term = document.querySelector("#admin-orden-busqueda")?.value.trim() || "";
  setAdminOrdenMessage("", "error");
  adminOrdenState.loading = true;
  renderAdminOrdenesList();
  try {
    const payload = await apiGet(`/api/admin/ordenes?busqueda=${encodeURIComponent(term)}`);
    adminOrdenState.ordenes = payload.data || [];
    adminOrdenState.ordenesLoaded = true;
    const selectedStillVisible = adminOrdenState.ordenes.some((orden) => orden.id === adminOrdenState.selectedOrden?.id);
    adminOrdenState.selectedOrden = selectedStillVisible ? adminOrdenState.selectedOrden : adminOrdenState.ordenes[0] || null;
    renderAdminOrdenesSection();
  } catch (error) {
    adminOrdenState.ordenesLoaded = true;
    setAdminOrdenMessage(error.message, "error");
  } finally {
    adminOrdenState.loading = false;
    renderAdminOrdenesList();
  }
}

async function cargarAdminEmpleados() {
  try {
    const payload = await apiGet("/api/creador-ordenes/empleados");
    adminOrdenState.empleados = payload.data || [];
    adminOrdenState.empleadosLoaded = true;
    renderAdminOrdenesSection();
  } catch {
    adminOrdenState.empleados = [];
    adminOrdenState.empleadosLoaded = true;
  }
}

function renderAdminOrdenesList() {
  const container = document.querySelector("#admin-ordenes-list");
  if (!container) return;
  if (adminOrdenState.loading) {
    container.innerHTML = `<div class="empty-inline">Cargando ordenes...</div>`;
    return;
  }
  if (!adminOrdenState.ordenes.length) {
    container.innerHTML = `<div class="empty-inline">No hay ordenes cargadas.</div>`;
    return;
  }
  container.innerHTML = adminOrdenState.ordenes.map((orden) => `
    <button class="result-item ${adminOrdenState.selectedOrden?.id === orden.id ? "selected-result" : ""}" type="button" data-admin-orden-id="${orden.id}">
      <strong>OT-${orden.numero} - ${orden.nombre_cliente}</strong>
      <span>${orden.placa_moto || "Sin placa"} - ${orden.estado} - ${orden.nombre_usuario_responsable || "Sin asignar"}</span>
    </button>
  `).join("");
  container.querySelectorAll("[data-admin-orden-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      const orden = adminOrdenState.ordenes.find((item) => item.id === button.dataset.adminOrdenId);
      adminOrdenState.selectedOrden = orden || null;
      renderAdminOrdenesSection();
      if (orden) await cargarAdminOrdenDetalle(orden.id);
    });
  });
}

async function cargarAdminOrdenDetalle(id) {
  setAdminOrdenMessage("", "error");
  try {
    const payload = await apiGet(`/api/admin/ordenes/${id}`);
    adminOrdenState.selectedOrden = payload.data || adminOrdenState.selectedOrden;
    renderAdminOrdenesSection();
  } catch (error) {
    setAdminOrdenMessage(error.message, "error");
  }
}

async function guardarAdminOrden() {
  const orden = adminOrdenState.selectedOrden;
  if (!orden) return;
  const payload = {
    usuario_responsable_id: document.querySelector("#admin-orden-empleado")?.value || "",
    kilometraje_ingreso: Number(document.querySelector("#admin-orden-kilometraje")?.value || 0),
    nivel_combustible_porcentaje: Number(document.querySelector("#admin-orden-combustible")?.value || 0),
    estado: document.querySelector("#admin-orden-estado")?.value || "abierta",
    tipo_servicio: document.querySelector("#admin-orden-tipo-servicio")?.value.trim() || "",
    descripcion_falla: document.querySelector("#admin-orden-descripcion")?.value.trim() || "",
    diagnostico: document.querySelector("#admin-orden-diagnostico")?.value.trim() || "",
    trabajos_realizados: document.querySelector("#admin-orden-trabajos")?.value.trim() || "",
    observaciones: document.querySelector("#admin-orden-observaciones")?.value.trim() || ""
  };
  if (!payload.descripcion_falla) {
    setAdminOrdenMessage("La descripcion de falla es obligatoria.", "error");
    return;
  }
  try {
    const response = await apiPatch(`/api/admin/ordenes/${orden.id}`, payload);
    adminOrdenState.selectedOrden = response.data;
    adminOrdenState.ordenes = adminOrdenState.ordenes.map((item) => item.id === response.data.id ? response.data : item);
    renderAdminOrdenesSection();
    setAdminOrdenMessage("Orden actualizada correctamente.", "success");
  } catch (error) {
    setAdminOrdenMessage(error.message, "error");
  }
}

function renderEstadoOptions(selected) {
  const estados = ["abierta", "en_diagnostico", "aprobacion_pendiente", "en_proceso", "lista_para_entrega", "cerrada", "cancelada"];
  return estados.map((estado) => `<option value="${estado}" ${selected === estado ? "selected" : ""}>${estado}</option>`).join("");
}

function renderAdminEmpleadoOptions(selected) {
  const base = [`<option value="">Sin asignar</option>`];
  const selectedInList = adminOrdenState.empleados.some((empleado) => empleado.id === selected);
  if (selected && !selectedInList) {
    const nombre = adminOrdenState.selectedOrden?.nombre_usuario_responsable || "Responsable actual";
    base.push(`<option value="${selected}" selected>${escapeHtml(nombre)}</option>`);
  }
  const empleados = adminOrdenState.empleados.map((empleado) => `<option value="${empleado.id}" ${selected === empleado.id ? "selected" : ""}>${empleado.nombre_completo} - ${empleado.usuario}</option>`);
  return base.concat(empleados).join("");
}

function setAdminOrdenMessage(message, type) {
  const error = document.querySelector("#admin-orden-error");
  const success = document.querySelector("#admin-orden-success");
  if (error) {
    error.textContent = type === "error" ? message : "";
    error.hidden = type !== "error" || !message;
  }
  if (success) {
    success.textContent = type === "success" ? message : "";
    success.hidden = type !== "success" || !message;
  }
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function escapeAttribute(value) {
  return escapeHtml(value);
}

function renderOrdenesSection() {
  const campos = ordenState.formato?.campos || defaultFormatoCampos;
  adminContent.innerHTML = `
    <div class="content-intro">
      <div>
        <p class="content-eyebrow">Recepcion de trabajo</p>
        <h2>Crear orden y asignarla a un empleado</h2>
        <p>Busca el cliente, selecciona su motocicleta, agrega items desde catalogo y asigna la orden al empleado responsable.</p>
      </div>
    </div>
    <section class="order-builder">
      <div class="order-column">
        <div class="work-panel">
          <div class="panel-toolbar"><div><h3>Cliente y motocicleta</h3><p>Busca por nombre, documento, telefono o placa.</p></div></div>
          <div class="panel-body">
            <label class="field compact-field"><span>Buscar cliente</span><div class="input-shell"><span class="input-icon" aria-hidden="true">@</span><input id="orden-cliente-busqueda" type="search" placeholder="Nombre, telefono o placa" /></div></label>
            <div class="action-row"><button class="primary-action" id="buscar-cliente-btn" type="button">Buscar</button></div>
            <div class="result-list" id="clientes-resultados"></div>
            <div class="selection-box" id="cliente-seleccionado"></div>
          </div>
        </div>
        <div class="work-panel">
          <div class="panel-toolbar"><div><h3>Datos de la orden</h3><p>Informacion de ingreso al taller.</p></div></div>
          <div class="panel-body form-grid">
            ${campos.map((campo) => renderDynamicField(campo, "orden", false)).join("")}
          </div>
        </div>
      </div>
      <div class="order-column">
        <div class="work-panel">
          <div class="panel-toolbar"><div><h3>Empleado responsable</h3><p>Selecciona quien recibira la orden asignada.</p></div></div>
          <div class="panel-body">
            <button class="secondary-light-button" id="cargar-empleados-btn" type="button">Cargar empleados</button>
            <select class="plain-input" id="orden-empleado"></select>
          </div>
        </div>
        <div class="work-panel">
          <div class="panel-toolbar"><div><h3>Items del trabajo</h3><p>Selecciona los items del catalogo que requiere la moto.</p></div></div>
          <div class="panel-body">
            <button class="secondary-light-button" id="agregar-item-orden-btn" type="button">Agregar item</button>
            <div class="selected-items" id="items-seleccionados"></div>
          </div>
        </div>
        <div class="create-order-footer">
          <p class="form-error" id="orden-error" hidden></p>
          <p class="form-success" id="orden-success" hidden></p>
          <button class="submit-button" id="crear-orden-btn" type="button">Crear orden de trabajo</button>
        </div>
      </div>
    </section>
  `;

  bindOrdenesEvents();
  renderClientesResultados();
  renderEmpleadoSelect();
  renderSelectedItems();
  if (!ordenState.formatoLoaded) {
    cargarFormatoCreador();
  }
  if (!ordenState.catalogoItems.length) {
    cargarCatalogoItems();
  }
}

async function cargarFormatoCreador() {
  try {
    const payload = await apiGet("/api/creador-ordenes/formato-orden");
    ordenState.formato = payload.data || { campos: defaultFormatoCampos };
  } catch {
    ordenState.formato = { campos: defaultFormatoCampos };
  } finally {
    ordenState.formatoLoaded = true;
    renderOrdenesSection();
  }
}

function renderDynamicField(campo, prefix, disabled) {
  const id = `${prefix}-field-${campo.id}`;
  const required = campo.requerido ? "required" : "";
  const disabledAttribute = disabled ? "disabled" : "";
  const spanClass = campo.tipo === "area" ? " span-2" : "";
  const label = `${escapeHtml(campo.label)}${campo.requerido ? " *" : ""}`;

  if (campo.tipo === "area") {
    return `<label class="field compact-field${spanClass}"><span>${label}</span><textarea class="plain-textarea" id="${id}" data-campo-id="${escapeAttribute(campo.id)}" data-campo-tipo="${campo.tipo}" ${required} ${disabledAttribute}></textarea></label>`;
  }
  if (campo.tipo === "checkbox") {
    return `<label class="field compact-field check-card"><span>${label}</span><input type="checkbox" id="${id}" data-campo-id="${escapeAttribute(campo.id)}" data-campo-tipo="${campo.tipo}" ${disabledAttribute} /></label>`;
  }
  const inputType = campo.tipo === "numero" ? "number" : campo.tipo === "fecha" ? "date" : "text";
  return `<label class="field compact-field${spanClass}"><span>${label}</span><input class="plain-input" id="${id}" type="${inputType}" data-campo-id="${escapeAttribute(campo.id)}" data-campo-tipo="${campo.tipo}" ${required} ${disabledAttribute} /></label>`;
}

function bindOrdenesEvents() {
  document.querySelector("#buscar-cliente-btn").addEventListener("click", buscarClientes);
  document.querySelector("#cargar-empleados-btn").addEventListener("click", cargarEmpleados);
  document.querySelector("#agregar-item-orden-btn").addEventListener("click", agregarItemOrden);
  document.querySelector("#crear-orden-btn").addEventListener("click", crearOrdenTrabajo);
}

async function buscarClientes() {
  const term = document.querySelector("#orden-cliente-busqueda").value.trim();
  setOrdenError("");
  if (term.length < 2) {
    setOrdenError("Escribe al menos 2 caracteres para buscar.");
    return;
  }

  try {
    const payload = await apiGet(`/api/creador-ordenes/clientes?busqueda=${encodeURIComponent(term)}`);
    ordenState.clientes = payload.data || [];
    renderClientesResultados();
  } catch (error) {
    setOrdenError(error.message);
  }
}

function renderClientesResultados() {
  const container = document.querySelector("#clientes-resultados");
  if (!container) return;
  if (!ordenState.clientes.length) {
    container.innerHTML = `<div class="empty-inline">Sin resultados cargados.</div>`;
    renderClienteSeleccionado();
    return;
  }
  container.innerHTML = ordenState.clientes.map((cliente) => `
    <button class="result-item" type="button" data-cliente-id="${cliente.id}">
      <strong>${cliente.nombre_completo}</strong>
      <span>${cliente.telefono || "Sin telefono"} - ${cliente.motos?.length || 0} moto(s)</span>
    </button>
  `).join("");
  container.querySelectorAll("[data-cliente-id]").forEach((button) => {
    button.addEventListener("click", () => {
      ordenState.selectedCliente = ordenState.clientes.find((cliente) => cliente.id === button.dataset.clienteId);
      ordenState.selectedMoto = ordenState.selectedCliente?.motos?.[0] || null;
      renderClienteSeleccionado();
    });
  });
  renderClienteSeleccionado();
}

function renderClienteSeleccionado() {
  const container = document.querySelector("#cliente-seleccionado");
  if (!container) return;
  const cliente = ordenState.selectedCliente;
  if (!cliente) {
    container.innerHTML = `<span>Selecciona un cliente para ver sus motocicletas.</span>`;
    return;
  }
  container.innerHTML = `
    <strong>${cliente.nombre_completo}</strong>
    <span>Telefono: ${cliente.telefono || "Sin telefono"}</span>
    <label class="field compact-field">
      <span>Motocicleta</span>
      <select class="plain-input" id="orden-moto">
        ${(cliente.motos || []).map((moto) => `<option value="${moto.id}" ${ordenState.selectedMoto?.id === moto.id ? "selected" : ""}>${moto.placa || "Sin placa"} - ${moto.marca} ${moto.modelo} - ${moto.kilometraje_actual} km</option>`).join("")}
      </select>
    </label>
  `;
  container.querySelector("#orden-moto")?.addEventListener("change", (event) => {
    ordenState.selectedMoto = cliente.motos.find((moto) => moto.id === event.target.value) || null;
  });
}

async function cargarEmpleados() {
  setOrdenError("");
  try {
    const payload = await apiGet("/api/creador-ordenes/empleados");
    ordenState.empleados = payload.data || [];
    ordenState.selectedEmpleado = ordenState.selectedEmpleado || ordenState.empleados[0]?.id || "";
    renderEmpleadoSelect();
  } catch (error) {
    setOrdenError(error.message);
  }
}

function renderEmpleadoSelect() {
  const select = document.querySelector("#orden-empleado");
  if (!select) return;
  if (!ordenState.empleados.length) {
    select.innerHTML = `<option value="">Sin empleados cargados</option>`;
    return;
  }
  select.innerHTML = ordenState.empleados.map((empleado) => `<option value="${empleado.id}" ${ordenState.selectedEmpleado === empleado.id ? "selected" : ""}>${empleado.nombre_completo} - ${empleado.usuario}</option>`).join("");
  select.addEventListener("change", () => {
    ordenState.selectedEmpleado = select.value;
  });
}

async function cargarCatalogoItems() {
  setOrdenError("");
  try {
    const payload = await apiGet("/api/creador-ordenes/catalogo-items?limit=100");
    ordenState.catalogoItems = payload.data || [];
    renderSelectedItems();
  } catch (error) {
    setOrdenError(error.message);
  }
}

function agregarItemOrden() {
  if (!ordenState.catalogoItems.length) {
    setOrdenError("Primero carga items en el catalogo de servicios.");
    return;
  }
  const item = ordenState.catalogoItems[0];
  ordenState.selectedItems.push({
    catalogo_item_trabajo_id: item.id,
    nombre: item.nombre,
    tipo_item: item.tipo_item,
    cantidad: 1,
    valor_unitario: Number(item.valor_base || 0),
    descuento: 0
  });
  renderSelectedItems();
}

function renderSelectedItems() {
  const container = document.querySelector("#items-seleccionados");
  if (!container) return;
  if (!ordenState.selectedItems.length) {
    container.innerHTML = `<div class="empty-inline">No hay items agregados. Usa el boton Agregar item.</div>`;
    return;
  }
  container.innerHTML = ordenState.selectedItems.map((item, index) => `
    <div class="selected-item">
      <select class="plain-input item-select" data-field="catalogo_item_trabajo_id" data-index="${index}">
        ${renderCatalogoItemOptions(item.catalogo_item_trabajo_id)}
      </select>
      <input class="mini-input" type="number" min="0.01" step="0.01" value="${item.cantidad}" data-field="cantidad" data-index="${index}" />
      <input class="mini-input" type="number" min="0" step="100" value="${item.valor_unitario}" data-field="valor_unitario" data-index="${index}" />
      <button class="remove-item-button" type="button" data-remove-index="${index}">Quitar</button>
    </div>
  `).join("");
  container.querySelectorAll("[data-field]").forEach((input) => {
    input.addEventListener("change", () => {
      const item = ordenState.selectedItems[Number(input.dataset.index)];
      if (input.dataset.field === "catalogo_item_trabajo_id") {
        const catalogoItem = ordenState.catalogoItems.find((catalogo) => catalogo.id === input.value);
        item.catalogo_item_trabajo_id = input.value;
        item.nombre = catalogoItem?.nombre || "";
        item.tipo_item = catalogoItem?.tipo_item || "";
        item.valor_unitario = Number(catalogoItem?.valor_base || 0);
        renderSelectedItems();
        return;
      }
      item[input.dataset.field] = Number(input.value || 0);
    });
  });
  container.querySelectorAll("[data-remove-index]").forEach((button) => {
    button.addEventListener("click", () => {
      ordenState.selectedItems.splice(Number(button.dataset.removeIndex), 1);
      renderSelectedItems();
    });
  });
}

function renderCatalogoItemOptions(selectedID) {
  if (!ordenState.catalogoItems.length) {
    return `<option value="">Sin items en catalogo</option>`;
  }
  return ordenState.catalogoItems.map((item) => {
    const value = Number(item.valor_base || 0).toLocaleString("es-CO");
    return `<option value="${item.id}" ${selectedID === item.id ? "selected" : ""}>${item.nombre} - ${item.tipo_item} - $${value}</option>`;
  }).join("");
}

async function crearOrdenTrabajo() {
  setOrdenError("");
  setOrdenSuccess("");
  const valoresFormato = collectFormatoValues("orden");
  const requiredMissing = getMissingRequiredFormatoValues(valoresFormato);
  if (requiredMissing.length) {
    setOrdenError(`Completa los campos obligatorios: ${requiredMissing.join(", ")}.`);
    return;
  }
  const payload = {
    cliente_id: ordenState.selectedCliente?.id || "",
    moto_id: ordenState.selectedMoto?.id || "",
    usuario_responsable_id: document.querySelector("#orden-empleado")?.value || ordenState.selectedEmpleado,
    kilometraje_ingreso: Number(valoresFormato.kilometraje_ingreso || 0),
    nivel_combustible_porcentaje: Number(valoresFormato.nivel_combustible_porcentaje || 0),
    tipo_servicio: String(valoresFormato.tipo_servicio || "").trim(),
    descripcion_falla: String(valoresFormato.descripcion_falla || "Orden registrada desde formato configurable").trim(),
    observaciones: String(valoresFormato.observaciones || "").trim(),
    campos_formato: valoresFormato,
    items: ordenState.selectedItems.map((item) => ({
      catalogo_item_trabajo_id: item.catalogo_item_trabajo_id,
      cantidad: item.cantidad,
      valor_unitario: item.valor_unitario,
      descuento: item.descuento
    }))
  };
  if (!payload.cliente_id || !payload.moto_id || !payload.usuario_responsable_id || !payload.items.length) {
    setOrdenError("Selecciona cliente, moto, empleado e items del trabajo.");
    return;
  }
  try {
    const response = await apiPost("/api/creador-ordenes/ordenes", payload);
    setOrdenSuccess(`Orden creada correctamente. Numero: ${response.data?.numero || response.data?.id}`);
    ordenState.selectedItems = [];
    renderSelectedItems();
  } catch (error) {
    setOrdenError(error.message);
  }
}

function getMissingRequiredFormatoValues(values) {
  const campos = ordenState.formato?.campos || defaultFormatoCampos;
  return campos
    .filter((campo) => campo.requerido)
    .filter((campo) => {
      const value = values[campo.id];
      if (campo.tipo === "checkbox") return value !== true;
      return String(value || "").trim() === "";
    })
    .map((campo) => campo.label);
}

function collectFormatoValues(prefix) {
  const values = {};
  document.querySelectorAll(`[id^="${prefix}-field-"]`).forEach((input) => {
    const id = input.dataset.campoId;
    if (!id) return;
    if (input.dataset.campoTipo === "checkbox") {
      values[id] = input.checked;
      return;
    }
    values[id] = input.value;
  });
  return values;
}

function setOrdenError(message) {
  const error = document.querySelector("#orden-error");
  const success = document.querySelector("#orden-success");
  if (success && message) success.hidden = true;
  if (!error) return;
  error.textContent = message;
  error.hidden = !message;
}

function setOrdenSuccess(message) {
  const error = document.querySelector("#orden-error");
  const success = document.querySelector("#orden-success");
  if (error && message) error.hidden = true;
  if (!success) return;
  success.textContent = message;
  success.hidden = !message;
}

clearSession();
renderSession(null);
