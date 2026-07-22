"use client";

import { useRouter } from "next/navigation";
import { clearSession } from "../lib/api";

const adminItems = [
  ["ordenes", "OT", "Ordenes de trabajo", "/admin/dashboard"],
  ["catalogos", "C", "Catalogos", "/admin/catalogos"],
  ["empleados", "E", "Empleados", "/admin/empleados"],
  ["clientes", "CL", "Clientes", "/admin/dashboard"],
  ["historial", "HM", "Historial de motocicletas", "/admin/dashboard"]
];

export default function AppShell({ children, user, activeSection, title, actionLabel, creatorMode = false, employeeMode = false }) {
  const router = useRouter();
  const items = creatorMode
    ? [["ordenes", "OT", "Ordenes de trabajo", "/creador/ordenes"]]
    : employeeMode
      ? [["ordenes", "OT", "Mis ordenes", "/empleado/ordenes"]]
      : adminItems;
  const displayName = [user?.nombres, user?.apellidos].filter(Boolean).join(" ") || user?.usuario || "Usuario";

  function logout() {
    clearSession();
    router.push("/login");
  }

  return (
    <main className={`admin-app ${creatorMode || employeeMode ? "creator-profile" : ""}`}>
      <aside className="admin-sidebar" aria-label="Navegacion principal">
        <div className="sidebar-brand">
          <img src="/logo-principal.jpg" alt="Tecnomotos HB" />
          <div>
            <strong>Tecnomotos HB</strong>
            <span>{employeeMode ? "Perfil empleado" : creatorMode ? "Perfil creador de ordenes" : "Panel administrador"}</span>
          </div>
        </div>

        <nav className="sidebar-nav">
          {items.map(([id, icon, label, href]) => (
            <button className={`nav-item ${activeSection === id ? "active" : ""}`} type="button" key={id} onClick={() => router.push(href)}>
              <span className="nav-icon" aria-hidden="true">{icon}</span>
              <span>{label}</span>
            </button>
          ))}
        </nav>

        <div className="sidebar-session">
          <span>{displayName}</span>
          <small>{user?.rol || "sin rol"}</small>
          <button className="logout-button" type="button" onClick={logout}>Cerrar sesion</button>
        </div>
      </aside>

      <section className="admin-main">
        <header className="admin-topbar">
          <div>
            <p className="brand-kicker">{employeeMode ? "Empleado" : creatorMode ? "Recepcion" : "Administracion"}</p>
            <h1>{title}</h1>
          </div>
          <div className="topbar-actions">
            <span className="status-pill">API conectada</span>
            <button className="primary-action" type="button">{actionLabel}</button>
          </div>
        </header>

        <section className="admin-content">{children}</section>
      </section>
    </main>
  );
}
