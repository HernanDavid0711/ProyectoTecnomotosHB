"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { login, saveSession } from "../lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [usuario, setUsuario] = useState("");
  const [contrasena, setContrasena] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event) {
    event.preventDefault();
    if (!usuario.trim() || !contrasena.trim()) {
      setError("Ingresa usuario y contrasena.");
      return;
    }

    setLoading(true);
    setError("");

    try {
      const payload = await login(usuario.trim(), contrasena.trim());
      saveSession(payload);
      const rol = payload.data?.rol;
      if (rol === "creador_ordenes") {
        router.push("/creador/ordenes");
      } else if (rol === "empleado") {
        router.push("/empleado/ordenes");
      } else {
        router.push("/admin/dashboard");
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="login-page">
      <section className="brand-panel" aria-label="Tecnomotos HB">
        <div className="brand-panel__inner">
          <img className="brand-logo" src="/logo-principal.jpg" alt="Tecnomotos HB" />
          <div className="brand-copy">
            <p className="brand-kicker">Sistema operativo del taller</p>
            <h1>Tecnomotos HB</h1>
            <p>Gestiona clientes, motos, usuarios y ordenes de trabajo desde un solo punto de control.</p>
          </div>
        </div>
      </section>

      <section className="login-panel" aria-label="Inicio de sesion">
        <form className="login-form" onSubmit={handleSubmit}>
          <div className="form-heading">
            <div className="form-icon" aria-hidden="true">HB</div>
            <div>
              <h2>Iniciar sesion</h2>
              <p>Ingresa con tu usuario autorizado.</p>
            </div>
          </div>

          <label className="field">
            <span>Usuario o correo</span>
            <div className="input-shell">
              <span className="input-icon" aria-hidden="true">@</span>
              <input
                value={usuario}
                onChange={(event) => setUsuario(event.target.value)}
                type="text"
                autoComplete="username"
                placeholder="admin"
                required
              />
            </div>
          </label>

          <label className="field">
            <span>Contrasena</span>
            <div className="input-shell">
              <span className="input-icon" aria-hidden="true">*</span>
              <input
                value={contrasena}
                onChange={(event) => setContrasena(event.target.value)}
                type={showPassword ? "text" : "password"}
                autoComplete="current-password"
                placeholder="Tu contrasena"
                required
              />
              <button
                className="icon-button"
                type="button"
                onClick={() => setShowPassword((value) => !value)}
                aria-label={showPassword ? "Ocultar contrasena" : "Mostrar contrasena"}
              >
                {showPassword ? "Ocultar" : "Ver"}
              </button>
            </div>
          </label>

          {error ? <p className="form-error">{error}</p> : null}

          <button className="submit-button" type="submit" disabled={loading}>
            {loading ? "Validando..." : "Entrar"}
          </button>
        </form>
      </section>
    </main>
  );
}
