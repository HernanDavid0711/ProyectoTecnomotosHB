"use client";

import { useEffect, useState } from "react";
import AppShell from "../../components/AppShell";
import useRequireSession from "../../hooks/useRequireSession";
import { apiGet } from "../../lib/api";

export default function AdminDashboardPage() {
  const { session, ready } = useRequireSession();
  const [busqueda, setBusqueda] = useState("");
  const [ordenes, setOrdenes] = useState([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (ready) cargarOrdenes("");
  }, [ready]);

  if (!ready) return null;

  async function cargarOrdenes(term = busqueda) {
    setError("");
    setLoading(true);
    try {
      const payload = await apiGet(`/api/admin/ordenes?busqueda=${encodeURIComponent(term.trim())}&limit=100`);
      setOrdenes(payload.data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function buscarOrdenes(event) {
    event.preventDefault();
    cargarOrdenes();
  }

  return (
    <AppShell user={session.user} activeSection="ordenes" title="Ordenes de trabajo" actionLabel="Buscar ordenes">
      <div className="content-intro">
        <div>
          <p className="content-eyebrow">Consulta operativa</p>
          <h2>Buscar ordenes de trabajo</h2>
          <p>Consulta las ordenes por nombre del cliente, placa, numero de orden o motocicleta.</p>
        </div>
      </div>

      <section className="work-panel admin-orders-search-panel">
        <form className="admin-orders-search" onSubmit={buscarOrdenes}>
          <label className="field compact-field">
            <span>Buscar</span>
            <div className="input-shell">
              <span className="input-icon" aria-hidden="true">@</span>
              <input value={busqueda} onChange={(event) => setBusqueda(event.target.value)} type="search" placeholder="Nombre, placa, moto o numero de orden" />
            </div>
          </label>
          <button className="submit-button" type="submit">{loading ? "Buscando..." : "Buscar"}</button>
        </form>
        {error ? <p className="form-error">{error}</p> : null}
      </section>

      <section className="work-panel admin-orders-results">
        <div className="panel-toolbar">
          <div>
            <h3>Resultados</h3>
            <p>{ordenes.length} orden(es) encontradas</p>
          </div>
        </div>
        <div className="data-table" role="table" aria-label="Ordenes de trabajo">
          <div className="table-row table-head admin-orders-row" role="row">
            <span>Orden</span>
            <span>Cliente</span>
            <span>Motocicleta</span>
            <span>Estado</span>
            <span>Responsable</span>
          </div>
          {ordenes.length ? ordenes.map((orden) => (
            <div className="table-row admin-orders-row" role="row" key={orden.id}>
              <div>
                <strong>OT-{orden.numero}</strong>
                <span>{formatearFecha(orden.fecha_ingreso)}</span>
              </div>
              <div>
                <strong>{orden.nombre_cliente || "Sin cliente"}</strong>
                <span>{orden.campos_formato?.numero_telefono || "Sin telefono"}</span>
              </div>
              <div>
                <strong>{orden.placa_moto || orden.campos_formato?.placa || "Sin placa"}</strong>
                <span>{orden.campos_formato?.moto || orden.campos_formato?.modelo || "Sin detalle"}</span>
              </div>
              <span className="state-pill is-active">{estadoLabel(orden.estado)}</span>
              <span>{orden.nombre_usuario_responsable || "Sin asignar"}</span>
            </div>
          )) : (
            <div className="empty-state">
              <strong>Sin ordenes para mostrar</strong>
              <span>Busca por cliente, placa o motocicleta.</span>
            </div>
          )}
        </div>
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

function formatearFecha(value) {
  if (!value) return "Sin fecha";
  return new Intl.DateTimeFormat("es-CO", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit"
  }).format(new Date(value));
}
