"use client";

import { useEffect, useMemo, useState } from "react";
import AppShell from "../../components/AppShell";
import useRequireSession from "../../hooks/useRequireSession";
import { apiGet, apiPost } from "../../lib/api";

export default function CreadorOrdenesPage() {
  const { session, ready } = useRequireSession("creador_ordenes");
  const [clientes, setClientes] = useState([]);
  const [clienteBusqueda, setClienteBusqueda] = useState("");
  const [placaBusqueda, setPlacaBusqueda] = useState("");
  const [selectedCliente, setSelectedCliente] = useState(null);
  const [selectedMotoID, setSelectedMotoID] = useState("");
  const [empleados, setEmpleados] = useState([]);
  const [empleadoID, setEmpleadoID] = useState("");
  const [catalogo, setCatalogo] = useState([]);
  const [items, setItems] = useState([]);
  const [kilometraje, setKilometraje] = useState("");
  const [observacionesMoto, setObservacionesMoto] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const selectedMoto = useMemo(
    () => selectedCliente?.motos?.find((moto) => moto.id === selectedMotoID) || null,
    [selectedCliente, selectedMotoID]
  );
  const totalItems = useMemo(
    () => items.reduce((total, item) => total + (Number(item.cantidad || 0) * Number(item.valor_unitario || 0)) - Number(item.descuento || 0), 0),
    [items]
  );

  useEffect(() => {
    if (!ready) return;
    apiGet("/api/creador-ordenes/catalogo-items?limit=100")
      .then((payload) => setCatalogo(payload.data || []))
      .catch((err) => setError(err.message));
    apiGet("/api/creador-ordenes/empleados")
      .then((payload) => {
        const data = payload.data || [];
        setEmpleados(data);
        setEmpleadoID(data[0]?.id || "");
      })
      .catch((err) => setError(err.message));
  }, [ready]);

  useEffect(() => {
    if (!ready) return;
    const term = clienteBusqueda.trim();
    if (term.length < 2) {
      setClientes([]);
      return;
    }

    const timeoutID = window.setTimeout(() => {
      buscarCoincidencias(term, "cliente");
    }, 350);

    return () => window.clearTimeout(timeoutID);
  }, [clienteBusqueda, ready]);

  useEffect(() => {
    if (!ready) return;
    const term = placaBusqueda.trim();
    if (term.length < 2) return;

    const timeoutID = window.setTimeout(() => {
      buscarCoincidencias(term, "placa");
    }, 350);

    return () => window.clearTimeout(timeoutID);
  }, [placaBusqueda, ready]);

  if (!ready) return null;

  async function buscarCoincidencias(term, origen) {
    setError("");
    try {
      const payload = await apiGet(`/api/creador-ordenes/clientes?busqueda=${encodeURIComponent(term)}&limit=10`);
      const data = payload.data || [];
      setClientes(data);
      autoseleccionarCoincidencia(data, term, origen);
    } catch (err) {
      setError(err.message);
    }
  }

  function autoseleccionarCoincidencia(data, term, origen) {
    if (!data.length) return;

    if (origen === "placa") {
      const normalizedTerm = term.toUpperCase().replaceAll(" ", "").replaceAll("-", "");
      for (const cliente of data) {
        const moto = cliente.motos?.find((item) => (item.placa || "").toUpperCase().includes(normalizedTerm));
        if (moto) {
          seleccionarCliente(cliente, moto.id, false);
          return;
        }
      }
    }

    const clienteConMoto = data.find((cliente) => cliente.motos?.length);
    if (clienteConMoto) {
      seleccionarCliente(clienteConMoto, clienteConMoto.motos[0].id, false);
    }
  }

  function seleccionarCliente(cliente, motoID = "", actualizarBusquedas = true) {
    setSelectedCliente(cliente);
    const moto = cliente.motos?.find((item) => item.id === motoID) || cliente.motos?.[0] || null;
    setSelectedMotoID(moto?.id || "");
    setKilometraje(moto?.kilometraje_actual ? String(moto.kilometraje_actual) : "");
    if (actualizarBusquedas) {
      setClienteBusqueda(cliente.nombre_completo || "");
      setPlacaBusqueda(moto?.placa || "");
    }
  }

  function seleccionarMoto(motoID) {
    setSelectedMotoID(motoID);
    const moto = selectedCliente?.motos?.find((item) => item.id === motoID) || null;
    setKilometraje(moto?.kilometraje_actual ? String(moto.kilometraje_actual) : "");
    setPlacaBusqueda(moto?.placa || "");
  }

  function agregarItem() {
    if (!catalogo.length) {
      setError("Primero carga items en el catalogo de servicios.");
      return;
    }
    setError("");
    setItems((current) => current.concat({
      catalogo_item_trabajo_id: "",
      item_busqueda: "",
      cantidad: 1,
      valor_unitario: "",
      descuento: 0
    }));
  }

  function actualizarItem(index, field, value) {
    setItems((current) => current.map((item, itemIndex) => {
      if (itemIndex !== index) return item;
      if (field === "item_busqueda") {
        return { ...item, item_busqueda: value, catalogo_item_trabajo_id: "" };
      }
      if (field === "valor_unitario") {
        return { ...item, valor_unitario: value };
      }
      return { ...item, [field]: Number(value || 0) };
    }));
  }

  function seleccionarCatalogoItem(index, catalogoItem) {
    setItems((current) => current.map((item, itemIndex) => {
      if (itemIndex !== index) return item;
      return {
        ...item,
        catalogo_item_trabajo_id: catalogoItem.id,
        item_busqueda: catalogoItem.nombre,
        valor_unitario: precioInicialItem(catalogoItem)
      };
    }));
  }

  function catalogoFiltrado(term) {
    const normalizedTerm = normalizarBusqueda(term);
    if (!normalizedTerm) return catalogo.slice(0, 8);
    return catalogo
      .filter((option) => normalizarBusqueda(`${option.nombre} ${option.tipo_item} ${option.descripcion || ""}`).includes(normalizedTerm))
      .slice(0, 8);
  }

  async function crearOrden() {
    setError("");
    setSuccess("");

    if (!selectedCliente || !selectedMotoID || !kilometraje) {
      setError("Selecciona cliente y motocicleta; valida que el kilometraje quede completo.");
      return;
    }
    if (!observacionesMoto.trim()) {
      setError("Agrega las observaciones de la motocicleta.");
      return;
    }
    if (!empleadoID) {
      setError("Selecciona el empleado responsable.");
      return;
    }
    if (!items.length) {
      setError("Agrega al menos un item de servicio a la orden.");
      return;
    }
    if (items.some((item) => !item.catalogo_item_trabajo_id || Number(item.cantidad || 0) <= 0 || Number(item.valor_unitario || 0) < 0)) {
      setError("Revisa los items: cada item debe tener servicio, cantidad y precio validos.");
      return;
    }

    try {
      const itemsNormalizados = items.map((item) => ({
        catalogo_item_trabajo_id: item.catalogo_item_trabajo_id,
        cantidad: Number(item.cantidad || 0),
        valor_unitario: Number(item.valor_unitario || 0),
        descuento: Number(item.descuento || 0)
      }));
      const camposFormato = {
        cliente: selectedCliente.nombre_completo,
        numero_telefono: selectedCliente.telefono || "",
        placa: selectedMoto?.placa || "",
        moto: selectedMoto ? `${selectedMoto.marca} ${selectedMoto.modelo}`.trim() : "",
        modelo: selectedMoto?.modelo || "",
        kilometraje_ingreso: kilometraje,
        observaciones_motocicleta: observacionesMoto
      };
      const payload = {
        cliente_id: selectedCliente.id,
        moto_id: selectedMotoID,
        usuario_responsable_id: empleadoID,
        kilometraje_ingreso: Number(kilometraje || 0),
        nivel_combustible_porcentaje: 0,
        tipo_servicio: "",
        descripcion_falla: observacionesMoto.trim(),
        observaciones: observacionesMoto.trim(),
        campos_formato: camposFormato,
        items: itemsNormalizados
      };
      const response = await apiPost("/api/creador-ordenes/ordenes", payload);
      setSuccess(`Orden creada correctamente. Numero: ${response.data?.numero || response.data?.id}`);
      setItems([]);
      setObservacionesMoto("");
      setKilometraje("");
      setSelectedCliente(null);
      setSelectedMotoID("");
      setClientes([]);
      setClienteBusqueda("");
      setPlacaBusqueda("");
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <AppShell user={session.user} activeSection="ordenes" title="Orden de trabajo" actionLabel="Crear orden" creatorMode>
      <div className="content-intro tablet-intro">
        <div>
          <p className="content-eyebrow">Recepcion en tablet</p>
          <h2>Formulario unico de ingreso</h2>
          <p>Completa los datos del cliente y la motocicleta para crear la orden de trabajo.</p>
        </div>
      </div>

      <section className="tablet-order-card">
        <header className="order-document-header">
          <div>
            <span>Formato operativo</span>
            <strong>Orden de trabajo</strong>
          </div>
          <div>
            <span>Estado inicial</span>
            <strong>abierta</strong>
          </div>
        </header>

        <div className="tablet-search-row">
          <label className="field compact-field">
            <span>Cliente</span>
            <div className="input-shell">
              <span className="input-icon" aria-hidden="true">@</span>
              <input value={clienteBusqueda} onChange={(event) => setClienteBusqueda(event.target.value)} type="search" placeholder="Escribe el nombre del cliente" />
            </div>
          </label>
          <label className="field compact-field">
            <span>Placa</span>
            <div className="input-shell">
              <span className="input-icon" aria-hidden="true">#</span>
              <input value={placaBusqueda} onChange={(event) => setPlacaBusqueda(event.target.value)} type="search" placeholder="Escribe la placa" />
            </div>
          </label>
        </div>

        {clientes.length ? (
          <div className="tablet-client-results">
            {clientes.map((cliente) => (
              <button className="result-item" type="button" key={cliente.id} onClick={() => seleccionarCliente(cliente)}>
                <strong>{cliente.nombre_completo}</strong>
                <span>{cliente.telefono || "Sin telefono"} - {(cliente.motos || []).map((moto) => moto.placa).filter(Boolean).join(", ") || "Sin motos"}</span>
              </button>
            ))}
          </div>
        ) : null}

        <div className="tablet-order-form">
          <label className="field compact-field">
            <span>Cliente</span>
            <input className="plain-input" value={selectedCliente?.nombre_completo || ""} readOnly placeholder="Selecciona un cliente" />
          </label>
          <label className="field compact-field">
            <span>Numero de telefono</span>
            <input className="plain-input" value={selectedCliente?.telefono || ""} readOnly placeholder="Telefono" />
          </label>
          <label className="field compact-field">
            <span>Moto</span>
            <select className="plain-input" value={selectedMotoID} onChange={(event) => seleccionarMoto(event.target.value)}>
              {selectedCliente?.motos?.length ? selectedCliente.motos.map((moto) => (
                <option value={moto.id} key={moto.id}>{moto.placa || "Sin placa"} - {moto.marca}</option>
              )) : <option value="">Selecciona cliente</option>}
            </select>
          </label>
          <label className="field compact-field">
            <span>Placa</span>
            <input className="plain-input" value={selectedMoto?.placa || ""} readOnly placeholder="Placa" />
          </label>
          <label className="field compact-field">
            <span>Modelo</span>
            <input className="plain-input" value={selectedMoto?.modelo || ""} readOnly placeholder="Modelo" />
          </label>
          <label className="field compact-field">
            <span>Kilometraje</span>
            <input className="plain-input" type="number" min="0" value={kilometraje} onChange={(event) => setKilometraje(event.target.value)} placeholder="Kilometraje actual" />
          </label>
          <label className="field compact-field tablet-span">
            <span>Observaciones de la motocicleta</span>
            <textarea className="plain-textarea" value={observacionesMoto} onChange={(event) => setObservacionesMoto(event.target.value)} placeholder="Describe observaciones, falla o solicitud del cliente" />
          </label>

          <label className="field compact-field">
            <span>Empleado responsable</span>
            <select className="plain-input" value={empleadoID} onChange={(event) => setEmpleadoID(event.target.value)}>
              {empleados.length ? empleados.map((empleado) => (
                <option value={empleado.id} key={empleado.id}>{empleado.nombre_completo} - {empleado.usuario}</option>
              )) : <option value="">Sin empleados cargados</option>}
            </select>
          </label>

          <div className="tablet-span tablet-items-box">
            <div className="panel-toolbar tablet-items-toolbar">
              <div>
                <h3>Items de servicio</h3>
                <p>Agrega los trabajos requeridos para la moto.</p>
              </div>
              <button className="secondary-light-button" type="button" onClick={agregarItem}>Agregar item</button>
            </div>
            <div className="selected-items">
              {items.length ? items.map((item, index) => (
                <div className="selected-item selected-item-autocomplete" key={index}>
                  <div className="item-autocomplete">
                    <input className="plain-input item-select" value={item.item_busqueda || ""} onChange={(event) => actualizarItem(index, "item_busqueda", event.target.value)} type="search" placeholder="Escribe el item de servicio" />
                    {item.item_busqueda && !item.catalogo_item_trabajo_id ? (
                      <div className="item-autocomplete-results">
                        {catalogoFiltrado(item.item_busqueda).length ? catalogoFiltrado(item.item_busqueda).map((option) => (
                          <button className="result-item" type="button" key={option.id} onClick={() => seleccionarCatalogoItem(index, option)}>
                            <strong>{option.nombre}</strong>
                            <span>{catalogoItemMeta(option)}</span>
                          </button>
                        )) : (
                          <div className="empty-inline">Sin coincidencias en catalogo.</div>
                        )}
                      </div>
                    ) : null}
                  </div>
                  <input className="mini-input" type="number" min="0.01" step="0.01" value={item.cantidad} onChange={(event) => actualizarItem(index, "cantidad", event.target.value)} />
                  <input className="mini-input" type="number" min="0" step="100" value={item.valor_unitario} onChange={(event) => actualizarItem(index, "valor_unitario", event.target.value)} placeholder="Precio" />
                  <button className="remove-item-button" type="button" onClick={() => setItems((current) => current.filter((_, itemIndex) => itemIndex !== index))}>Quitar</button>
                </div>
              )) : <div className="empty-inline">No hay items agregados.</div>}
            </div>
            <div className="tablet-total-row">
              <span>Total estimado</span>
              <strong>${totalItems.toLocaleString("es-CO")}</strong>
            </div>
          </div>
        </div>

        <footer className="tablet-form-footer">
          <div>
            {error ? <p className="form-error">{error}</p> : null}
            {success ? <p className="form-success">{success}</p> : null}
          </div>
          <button className="submit-button tablet-submit" type="button" onClick={crearOrden}>Crear orden de trabajo</button>
        </footer>
      </section>
    </AppShell>
  );
}

function catalogoItemLabel(option) {
  const valorBase = Number(option.valor_base || 0);
  const precio = valorBase > 0 ? ` - $${valorBase.toLocaleString("es-CO")}` : "";
  return `${option.nombre} - ${option.tipo_item}${precio}`;
}

function catalogoItemMeta(option) {
  const valorBase = Number(option.valor_base || 0);
  const precio = valorBase > 0 ? ` - $${valorBase.toLocaleString("es-CO")}` : "";
  return `${option.tipo_item}${precio}`;
}

function precioInicialItem(item) {
  const valorBase = Number(item?.valor_base || 0);
  return valorBase > 0 ? valorBase : "";
}

function normalizarBusqueda(value) {
  return String(value || "").trim().toLowerCase();
}
