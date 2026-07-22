"use client";

import { useEffect, useState } from "react";
import AppShell from "../../components/AppShell";
import useRequireSession from "../../hooks/useRequireSession";
import { apiGet } from "../../lib/api";

export default function EmpleadoOrdenesPage() {
  const { session, ready } = useRequireSession("empleado");
  const [ordenes, setOrdenes] = useState([]);
  const [estado, setEstado] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (ready) cargarOrdenes("");
  }, [ready]);

  if (!ready) return null;

  async function cargarOrdenes(estadoFiltro = estado) {
    setError("");
    setLoading(true);
    try {
      const payload = await apiGet(`/api/empleado/ordenes?estado=${encodeURIComponent(estadoFiltro)}&limit=100`);
      setOrdenes(payload.data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function cambiarEstado(event) {
    const value = event.target.value;
    setEstado(value);
    cargarOrdenes(value);
  }

  return (
    <AppShell user={session.user} activeSection="ordenes" title="Mis ordenes" actionLabel="Ordenes asignadas" employeeMode>
      <div className="content-intro">
        <div>
          <p className="content-eyebrow">Trabajo asignado</p>
          <h2>Ordenes de trabajo</h2>
          <p>Consulta desde el celular las ordenes que tienes asignadas.</p>
        </div>
      </div>

      <section className="employee-mobile-toolbar">
        <label className="field compact-field">
          <span>Estado</span>
          <select className="plain-input" value={estado} onChange={cambiarEstado}>
            <option value="">Todas</option>
            <option value="abierta">Abierta</option>
            <option value="en_diagnostico">En diagnostico</option>
            <option value="en_proceso">En proceso</option>
            <option value="lista_para_entrega">Lista para entrega</option>
            <option value="cerrada">Cerrada</option>
          </select>
        </label>
        <button className="secondary-light-button" type="button" onClick={() => cargarOrdenes()}>{loading ? "Cargando..." : "Actualizar"}</button>
      </section>

      {error ? <p className="form-error">{error}</p> : null}

      <section className="employee-order-list">
        {ordenes.length ? ordenes.map((orden) => (
          <article className="work-panel employee-order-card" key={orden.id}>
            <div className="employee-order-head">
              <div>
                <span>OT-{orden.numero}</span>
                <strong>{orden.nombre_cliente || "Sin cliente"}</strong>
              </div>
              <span className="state-pill is-active">{estadoLabel(orden.estado)}</span>
            </div>
            <div className="employee-order-grid">
              <span>Placa</span>
              <strong>{orden.placa_moto || orden.campos_formato?.placa || "Sin placa"}</strong>
              <span>Moto</span>
              <strong>{orden.campos_formato?.moto || orden.campos_formato?.modelo || "Sin detalle"}</strong>
              <span>Kilometraje</span>
              <strong>{Number(orden.kilometraje_ingreso || 0).toLocaleString("es-CO")} km</strong>
              <span>Observaciones</span>
              <strong>{orden.observaciones || orden.descripcion_falla || "Sin observaciones"}</strong>
            </div>
          </article>
        )) : (
          <div className="empty-state">
            <strong>Sin ordenes asignadas</strong>
            <span>Cuando admin asigne una orden, aparecera aqui.</span>
          </div>
        )}
      </section>
    </AppShell>
  );
}

function estadoLabel(value) {
  const labels = {
    abierta: "Abierta",
    en_diagnostico: "En diagnostico",
    aprobacion_pendiente: "Aprobacion pendiente",
    en_proceso: "En proceso",
    lista_para_entrega: "Lista para entrega",
    cerrada: "Cerrada",
    cancelada: "Cancelada"
  };
  return labels[value] || value || "Sin estado";
}
