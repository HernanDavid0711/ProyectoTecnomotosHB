"use client";

import { useEffect, useMemo, useState } from "react";
import AppShell from "../../components/AppShell";
import useRequireSession from "../../hooks/useRequireSession";
import { apiDelete, apiGet, apiPost, apiPut } from "../../lib/api";

const emptyForm = {
  usuario: "",
  nombres: "",
  apellidos: "",
  correo: "",
  telefono: "",
  contrasena: "",
  activo: true
};

export default function AdminEmpleadosPage() {
  const { session, ready } = useRequireSession();
  const [usuarios, setUsuarios] = useState([]);
  const [busqueda, setBusqueda] = useState("");
  const [form, setForm] = useState(emptyForm);
  const [editingID, setEditingID] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  const empleados = useMemo(() => usuarios.filter((usuario) => usuario.rol === "empleado"), [usuarios]);

  useEffect(() => {
    if (ready) cargarEmpleados("");
  }, [ready]);

  if (!ready) return null;

  async function cargarEmpleados(term = busqueda) {
    setError("");
    setLoading(true);
    try {
      const payload = await apiGet(`/api/admin/usuarios?busqueda=${encodeURIComponent(term.trim())}&limit=100`);
      setUsuarios(payload.data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function limpiarFormulario() {
    setEditingID("");
    setForm(emptyForm);
  }

  function editarEmpleado(empleado) {
    setEditingID(empleado.id);
    setForm({
      usuario: empleado.usuario || "",
      nombres: empleado.nombres || "",
      apellidos: empleado.apellidos || "",
      correo: empleado.correo || "",
      telefono: empleado.telefono || "",
      contrasena: "",
      activo: empleado.activo
    });
    setError("");
    setSuccess("");
  }

  async function guardarEmpleado(event) {
    event.preventDefault();
    setError("");
    setSuccess("");

    if (!form.usuario.trim() || !form.nombres.trim() || !form.apellidos.trim() || !form.correo.trim()) {
      setError("Usuario, nombres, apellidos y correo son obligatorios.");
      return;
    }
    if (!editingID && form.contrasena.trim().length < 8) {
      setError("La contrasena inicial debe tener minimo 8 caracteres.");
      return;
    }
    if (editingID && form.contrasena.trim() && form.contrasena.trim().length < 8) {
      setError("La nueva contrasena debe tener minimo 8 caracteres.");
      return;
    }

    try {
      const payload = {
        rol: "empleado",
        usuario: form.usuario.trim(),
        nombres: form.nombres.trim(),
        apellidos: form.apellidos.trim(),
        correo: form.correo.trim(),
        telefono: form.telefono.trim(),
        contrasena: form.contrasena.trim()
      };

      if (editingID) {
        await apiPut(`/api/admin/usuarios/${editingID}`, { ...payload, activo: form.activo });
        setSuccess("Empleado actualizado correctamente.");
      } else {
        await apiPost("/api/admin/usuarios", payload);
        setSuccess("Empleado creado correctamente.");
      }

      limpiarFormulario();
      await cargarEmpleados();
    } catch (err) {
      setError(err.message);
    }
  }

  async function desactivarEmpleado(id) {
    setError("");
    setSuccess("");
    try {
      await apiDelete(`/api/admin/usuarios/${id}`);
      setSuccess("Empleado desactivado correctamente.");
      await cargarEmpleados();
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <AppShell user={session.user} activeSection="empleados" title="Empleados" actionLabel={editingID ? "Editando empleado" : "Nuevo empleado"}>
      <div className="content-intro">
        <div>
          <p className="content-eyebrow">Equipo del taller</p>
          <h2>Gestionar empleados</h2>
          <p>Crea empleados con usuario y contrasena inicial para que ingresen desde celular a sus ordenes asignadas.</p>
        </div>
      </div>

      <section className="employee-admin-layout">
        <form className="work-panel employee-form" onSubmit={guardarEmpleado}>
          <div className="panel-toolbar">
            <div>
              <h3>{editingID ? "Editar empleado" : "Crear empleado"}</h3>
              <p>El usuario creado tendra rol empleado.</p>
            </div>
          </div>
          <div className="panel-body employee-form-grid">
            <label className="field compact-field">
              <span>Usuario</span>
              <input className="plain-input" value={form.usuario} onChange={(event) => setForm((current) => ({ ...current, usuario: event.target.value }))} placeholder="ej. mecanico1" />
            </label>
            <label className="field compact-field">
              <span>Contrasena inicial</span>
              <input className="plain-input" type="password" value={form.contrasena} onChange={(event) => setForm((current) => ({ ...current, contrasena: event.target.value }))} placeholder={editingID ? "Dejar vacia para conservar" : "Minimo 8 caracteres"} />
            </label>
            <label className="field compact-field">
              <span>Nombres</span>
              <input className="plain-input" value={form.nombres} onChange={(event) => setForm((current) => ({ ...current, nombres: event.target.value }))} placeholder="Nombres" />
            </label>
            <label className="field compact-field">
              <span>Apellidos</span>
              <input className="plain-input" value={form.apellidos} onChange={(event) => setForm((current) => ({ ...current, apellidos: event.target.value }))} placeholder="Apellidos" />
            </label>
            <label className="field compact-field">
              <span>Correo</span>
              <input className="plain-input" type="email" value={form.correo} onChange={(event) => setForm((current) => ({ ...current, correo: event.target.value }))} placeholder="empleado@correo.com" />
            </label>
            <label className="field compact-field">
              <span>Telefono</span>
              <input className="plain-input" value={form.telefono} onChange={(event) => setForm((current) => ({ ...current, telefono: event.target.value }))} placeholder="Telefono" />
            </label>
            {editingID ? (
              <label className="field compact-field check-card">
                <span>Activo</span>
                <input type="checkbox" checked={form.activo} onChange={(event) => setForm((current) => ({ ...current, activo: event.target.checked }))} />
              </label>
            ) : null}
            {error ? <p className="form-error">{error}</p> : null}
            {success ? <p className="form-success">{success}</p> : null}
            <div className="catalog-form-actions">
              <button className="submit-button" type="submit">{editingID ? "Guardar cambios" : "Crear empleado"}</button>
              {editingID ? <button className="secondary-light-button" type="button" onClick={limpiarFormulario}>Cancelar</button> : null}
            </div>
          </div>
        </form>

        <section className="work-panel employee-list-panel">
          <div className="panel-toolbar">
            <div>
              <h3>Empleados registrados</h3>
              <p>{empleados.length} empleado(s) disponibles para asignar ordenes.</p>
            </div>
            <form className="catalog-search" onSubmit={(event) => { event.preventDefault(); cargarEmpleados(); }}>
              <input className="search-input" value={busqueda} onChange={(event) => setBusqueda(event.target.value)} placeholder="Buscar empleado" />
              <button className="secondary-light-button" type="submit">{loading ? "Buscando..." : "Buscar"}</button>
            </form>
          </div>
          <div className="data-table" role="table" aria-label="Empleados">
            <div className="table-row table-head employee-row" role="row">
              <span>Empleado</span>
              <span>Usuario</span>
              <span>Contacto</span>
              <span>Estado</span>
              <span>Acciones</span>
            </div>
            {empleados.length ? empleados.map((empleado) => (
              <div className="table-row employee-row" role="row" key={empleado.id}>
                <div>
                  <strong>{empleado.nombres} {empleado.apellidos}</strong>
                  <span>{empleado.rol}</span>
                </div>
                <strong>{empleado.usuario}</strong>
                <div>
                  <span>{empleado.correo}</span>
                  <span>{empleado.telefono || "Sin telefono"}</span>
                </div>
                <span className={`state-pill ${empleado.activo ? "is-active" : "is-inactive"}`}>{empleado.activo ? "Activo" : "Inactivo"}</span>
                <div className="row-actions">
                  <button className="secondary-light-button" type="button" onClick={() => editarEmpleado(empleado)}>Editar</button>
                  {empleado.activo ? <button className="remove-item-button" type="button" onClick={() => desactivarEmpleado(empleado.id)}>Desactivar</button> : null}
                </div>
              </div>
            )) : (
              <div className="empty-state">
                <strong>Sin empleados registrados</strong>
                <span>Crea el primer empleado para asignarle ordenes de trabajo.</span>
              </div>
            )}
          </div>
        </section>
      </section>
    </AppShell>
  );
}
