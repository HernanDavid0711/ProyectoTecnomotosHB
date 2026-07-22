"use client";

import { useEffect, useState } from "react";
import AppShell from "../../components/AppShell";
import useRequireSession from "../../hooks/useRequireSession";
import { apiDelete, apiGet, apiPost, apiPut } from "../../lib/api";

const emptyItemForm = {
  tipo_item: "mano_obra",
  nombre: "",
  descripcion: "",
  activo: true
};

const emptyClienteForm = {
  nombres: "",
  correo: "",
  telefono: ""
};

const emptyMotoForm = {
  placa: "",
  marca: "",
  modelo: "",
  cilindraje: "",
  anio: "",
  color: "",
  kilometraje_actual: "",
  cliente_id: ""
};

const tipos = [
  ["mano_obra", "Mano de obra"],
  ["repuesto", "Repuesto"],
  ["servicio_externo", "Servicio externo"],
  ["insumo", "Insumo"]
];

export default function AdminCatalogosPage() {
  const { session, ready } = useRequireSession();
  const [items, setItems] = useState([]);
  const [clientes, setClientes] = useState([]);
  const [motos, setMotos] = useState([]);
  const [busquedaItems, setBusquedaItems] = useState("");
  const [busquedaClientes, setBusquedaClientes] = useState("");
  const [busquedaMotos, setBusquedaMotos] = useState("");
  const [itemForm, setItemForm] = useState(emptyItemForm);
  const [clienteForm, setClienteForm] = useState(emptyClienteForm);
  const [motoForm, setMotoForm] = useState(emptyMotoForm);
  const [catalogoActivo, setCatalogoActivo] = useState("");
  const [editingItemID, setEditingItemID] = useState("");
  const [editingClienteID, setEditingClienteID] = useState("");
  const [editingMotoID, setEditingMotoID] = useState("");
  const [itemError, setItemError] = useState("");
  const [itemSuccess, setItemSuccess] = useState("");
  const [clienteError, setClienteError] = useState("");
  const [clienteSuccess, setClienteSuccess] = useState("");
  const [motoError, setMotoError] = useState("");
  const [motoSuccess, setMotoSuccess] = useState("");

  useEffect(() => {
    if (ready) {
      cargarItems();
      cargarClientes();
      cargarMotos();
    }
  }, [ready]);

  if (!ready) return null;

  async function cargarItems() {
    setItemError("");
    try {
      const payload = await apiGet(`/api/admin/catalogo-items?busqueda=${encodeURIComponent(busquedaItems)}&limit=100`);
      setItems(payload.data || []);
    } catch (err) {
      setItemError(err.message);
    }
  }

  async function cargarClientes() {
    setClienteError("");
    try {
      const payload = await apiGet(`/api/clientes?busqueda=${encodeURIComponent(busquedaClientes)}&limit=100`);
      setClientes(payload.data || []);
    } catch (err) {
      setClienteError(err.message);
    }
  }

  async function cargarMotos() {
    setMotoError("");
    try {
      const payload = await apiGet(`/api/motos?busqueda=${encodeURIComponent(busquedaMotos)}&limit=100`);
      setMotos(payload.data || []);
    } catch (err) {
      setMotoError(err.message);
    }
  }

  function editarItem(item) {
    setEditingItemID(item.id);
    setItemForm({
      tipo_item: item.tipo_item,
      nombre: item.nombre,
      descripcion: item.descripcion || "",
      activo: item.activo
    });
    setItemSuccess("");
    setItemError("");
  }

  function editarCliente(cliente) {
    setEditingClienteID(cliente.id);
    setClienteForm({
      nombres: nombreCliente(cliente),
      correo: cliente.correo || "",
      telefono: cliente.telefono || ""
    });
    setClienteSuccess("");
    setClienteError("");
  }

  function editarMoto(moto) {
    setEditingMotoID(moto.id);
    setMotoForm({
      placa: moto.placa || "",
      marca: moto.marca || "",
      modelo: moto.modelo || "",
      cilindraje: String(moto.cilindraje || ""),
      anio: moto.anio ? String(moto.anio) : "",
      color: moto.color || "",
      kilometraje_actual: String(moto.kilometraje_actual || ""),
      cliente_id: moto.cliente_id || ""
    });
    setMotoSuccess("");
    setMotoError("");
  }

  function limpiarItemForm() {
    setEditingItemID("");
    setItemForm(emptyItemForm);
  }

  function limpiarClienteForm() {
    setEditingClienteID("");
    setClienteForm(emptyClienteForm);
  }

  function limpiarMotoForm() {
    setEditingMotoID("");
    setMotoForm(emptyMotoForm);
  }

  async function guardarItem(event) {
    event.preventDefault();
    setItemError("");
    setItemSuccess("");
    if (!itemForm.nombre.trim()) {
      setItemError("El nombre del item es obligatorio.");
      return;
    }

    try {
      const payload = {
        tipo_item: itemForm.tipo_item,
        nombre: itemForm.nombre.trim(),
        descripcion: itemForm.descripcion.trim(),
        activo: itemForm.activo
      };
      if (editingItemID) {
        await apiPut(`/api/admin/catalogo-items/${editingItemID}`, payload);
        setItemSuccess("Item actualizado correctamente.");
      } else {
        await apiPost("/api/admin/catalogo-items", payload);
        setItemSuccess("Item creado correctamente.");
      }
      limpiarItemForm();
      await cargarItems();
    } catch (err) {
      setItemError(err.message);
    }
  }

  async function guardarCliente(event) {
    event.preventDefault();
    setClienteError("");
    setClienteSuccess("");
    if (!clienteForm.nombres.trim()) {
      setClienteError("El nombre del cliente es obligatorio.");
      return;
    }

    try {
      const payload = {
        nombres: clienteForm.nombres.trim(),
        apellidos: "",
        telefono: clienteForm.telefono.trim(),
        correo: clienteForm.correo.trim(),
        direccion: "",
        tipo_documento: "",
        numero_documento: "",
        notas: ""
      };
      if (editingClienteID) {
        await apiPut(`/api/clientes/${editingClienteID}`, payload);
        setClienteSuccess("Cliente actualizado correctamente.");
      } else {
        await apiPost("/api/clientes", payload);
        setClienteSuccess("Cliente creado correctamente.");
      }
      limpiarClienteForm();
      await cargarClientes();
    } catch (err) {
      setClienteError(err.message);
    }
  }

  async function guardarMoto(event) {
    event.preventDefault();
    setMotoError("");
    setMotoSuccess("");
    if (!motoForm.placa.trim() || !motoForm.marca.trim() || !motoForm.modelo.trim() || !motoForm.cilindraje || !motoForm.cliente_id) {
      setMotoError("Placa, marca, modelo, cilindraje y cliente son obligatorios.");
      return;
    }

    try {
      const payload = {
        placa: motoForm.placa.trim(),
        marca: motoForm.marca.trim(),
        modelo: motoForm.modelo.trim(),
        cilindraje: Number(motoForm.cilindraje),
        anio: Number(motoForm.anio || 0),
        color: motoForm.color.trim(),
        kilometraje_actual: Number(motoForm.kilometraje_actual || 0),
        cliente_id: motoForm.cliente_id
      };
      if (editingMotoID) {
        await apiPut(`/api/motos/${editingMotoID}`, payload);
        setMotoSuccess("Motocicleta actualizada correctamente.");
      } else {
        await apiPost("/api/motos", payload);
        setMotoSuccess("Motocicleta creada correctamente.");
      }
      limpiarMotoForm();
      await cargarMotos();
    } catch (err) {
      setMotoError(err.message);
    }
  }

  async function desactivarItem(id) {
    setItemError("");
    setItemSuccess("");
    try {
      await apiDelete(`/api/admin/catalogo-items/${id}`);
      setItemSuccess("Item desactivado correctamente.");
      await cargarItems();
    } catch (err) {
      setItemError(err.message);
    }
  }

  async function desactivarCliente(id) {
    setClienteError("");
    setClienteSuccess("");
    try {
      await apiDelete(`/api/clientes/${id}`);
      setClienteSuccess("Cliente desactivado correctamente.");
      await cargarClientes();
    } catch (err) {
      setClienteError(err.message);
    }
  }

  return (
    <AppShell user={session.user} activeSection="catalogos" title="Catalogos" actionLabel="Catalogos">
      <div className="content-intro">
        <div>
          <p className="content-eyebrow">Catalogos operativos</p>
          <h2>{catalogoActivo ? catalogoTitulo(catalogoActivo) : "Catalogos del taller"}</h2>
          <p>{catalogoActivo ? "Gestiona los registros del catalogo seleccionado." : "Selecciona el catalogo que quieres administrar."}</p>
        </div>
        {catalogoActivo ? (
          <button className="secondary-light-button" type="button" onClick={() => setCatalogoActivo("")}>
            Volver a catalogos
          </button>
        ) : null}
      </div>

      {!catalogoActivo ? (
        <section className="catalog-menu-grid">
          <button className="work-panel catalog-menu-card" type="button" onClick={() => setCatalogoActivo("items")}>
            <span className="catalog-menu-icon">IT</span>
            <strong>Items Orden de Trabajo</strong>
            <small>{items.length} item(s) registrados</small>
          </button>
          <button className="work-panel catalog-menu-card" type="button" onClick={() => setCatalogoActivo("clientes")}>
            <span className="catalog-menu-icon">CL</span>
            <strong>Clientes</strong>
            <small>{clientes.length} cliente(s) registrados</small>
          </button>
          <button className="work-panel catalog-menu-card" type="button" onClick={() => setCatalogoActivo("motos")}>
            <span className="catalog-menu-icon">MT</span>
            <strong>Motocicletas</strong>
            <small>{motos.length} moto(s) registradas</small>
          </button>
        </section>
      ) : null}

      {catalogoActivo === "items" ? (
        <section className="catalog-detail-layout">
          <article className="work-panel catalog-card">
          <div className="panel-toolbar">
            <div>
              <h3>Items orden de trabajo</h3>
              <p>Trabajos, repuestos, insumos y servicios externos disponibles.</p>
            </div>
          </div>
          <form className="catalog-form catalog-card-form" onSubmit={guardarItem}>
            <div className="panel-body catalog-card-body">
              <label className="field compact-field">
                <span>Tipo de item</span>
                <select className="plain-input" value={itemForm.tipo_item} onChange={(event) => setItemForm((current) => ({ ...current, tipo_item: event.target.value }))}>
                  {tipos.map(([value, label]) => <option value={value} key={value}>{label}</option>)}
                </select>
              </label>
              <label className="field compact-field">
                <span>Nombre del item</span>
                <input className="plain-input" value={itemForm.nombre} onChange={(event) => setItemForm((current) => ({ ...current, nombre: event.target.value }))} placeholder="Ej. Cambio de aceite" />
              </label>
              <label className="field compact-field">
                <span>Descripcion</span>
                <textarea className="plain-textarea" value={itemForm.descripcion} onChange={(event) => setItemForm((current) => ({ ...current, descripcion: event.target.value }))} placeholder="Detalle del trabajo o insumo" />
              </label>
              {editingItemID ? (
                <label className="field compact-field check-card">
                  <span>Activo</span>
                  <input type="checkbox" checked={itemForm.activo} onChange={(event) => setItemForm((current) => ({ ...current, activo: event.target.checked }))} />
                </label>
              ) : null}
              {itemError ? <p className="form-error">{itemError}</p> : null}
              {itemSuccess ? <p className="form-success">{itemSuccess}</p> : null}
              <div className="catalog-form-actions">
                <button className="submit-button" type="submit">{editingItemID ? "Guardar cambios" : "Crear item"}</button>
                {editingItemID ? <button className="secondary-light-button" type="button" onClick={limpiarItemForm}>Cancelar</button> : null}
              </div>
            </div>
          </form>
          <div className="catalog-card-list">
            <div className="catalog-search">
              <input className="search-input" value={busquedaItems} onChange={(event) => setBusquedaItems(event.target.value)} placeholder="Buscar item" />
              <button className="secondary-light-button" type="button" onClick={cargarItems}>Buscar</button>
            </div>
            <CatalogItemsTable items={items} editarItem={editarItem} desactivarItem={desactivarItem} />
          </div>
          </article>
        </section>
      ) : null}

      {catalogoActivo === "clientes" ? (
        <section className="catalog-detail-layout">
          <article className="work-panel catalog-card">
          <div className="panel-toolbar">
            <div>
              <h3>Clientes</h3>
              <p>Contactos disponibles para las ordenes de trabajo.</p>
            </div>
          </div>
          <form className="catalog-form catalog-card-form" onSubmit={guardarCliente}>
            <div className="panel-body catalog-card-body">
              <label className="field compact-field">
                <span>Nombre</span>
                <input className="plain-input" value={clienteForm.nombres} onChange={(event) => setClienteForm((current) => ({ ...current, nombres: event.target.value }))} placeholder="Nombre del cliente" />
              </label>
              <label className="field compact-field">
                <span>Correo</span>
                <input className="plain-input" type="email" value={clienteForm.correo} onChange={(event) => setClienteForm((current) => ({ ...current, correo: event.target.value }))} placeholder="cliente@correo.com" />
              </label>
              <label className="field compact-field">
                <span>Numero de telefono</span>
                <input className="plain-input" value={clienteForm.telefono} onChange={(event) => setClienteForm((current) => ({ ...current, telefono: event.target.value }))} placeholder="Telefono" />
              </label>
              {clienteError ? <p className="form-error">{clienteError}</p> : null}
              {clienteSuccess ? <p className="form-success">{clienteSuccess}</p> : null}
              <div className="catalog-form-actions">
                <button className="submit-button" type="submit">{editingClienteID ? "Guardar cambios" : "Crear cliente"}</button>
                {editingClienteID ? <button className="secondary-light-button" type="button" onClick={limpiarClienteForm}>Cancelar</button> : null}
              </div>
            </div>
          </form>
          <div className="catalog-card-list">
            <div className="catalog-search">
              <input className="search-input" value={busquedaClientes} onChange={(event) => setBusquedaClientes(event.target.value)} placeholder="Buscar cliente" />
              <button className="secondary-light-button" type="button" onClick={cargarClientes}>Buscar</button>
            </div>
            <CatalogClientesTable clientes={clientes} editarCliente={editarCliente} desactivarCliente={desactivarCliente} />
          </div>
          </article>
        </section>
      ) : null}

      {catalogoActivo === "motos" ? (
        <section className="catalog-detail-layout">
          <article className="work-panel catalog-card">
          <div className="panel-toolbar">
            <div>
              <h3>Motocicletas</h3>
              <p>Motos registradas y asociadas a clientes del taller.</p>
            </div>
          </div>
          <form className="catalog-form catalog-card-form" onSubmit={guardarMoto}>
            <div className="panel-body catalog-card-body">
              <label className="field compact-field">
                <span>Cliente</span>
                <select className="plain-input" value={motoForm.cliente_id} onChange={(event) => setMotoForm((current) => ({ ...current, cliente_id: event.target.value }))}>
                  <option value="">Seleccionar cliente</option>
                  {clientes.map((cliente) => (
                    <option value={cliente.id} key={cliente.id}>{nombreCliente(cliente)}</option>
                  ))}
                </select>
              </label>
              <label className="field compact-field">
                <span>Placa</span>
                <input className="plain-input" value={motoForm.placa} onChange={(event) => setMotoForm((current) => ({ ...current, placa: event.target.value }))} placeholder="ABC123" />
              </label>
              <label className="field compact-field">
                <span>Marca</span>
                <input className="plain-input" value={motoForm.marca} onChange={(event) => setMotoForm((current) => ({ ...current, marca: event.target.value }))} placeholder="Yamaha" />
              </label>
              <label className="field compact-field">
                <span>Modelo</span>
                <input className="plain-input" value={motoForm.modelo} onChange={(event) => setMotoForm((current) => ({ ...current, modelo: event.target.value }))} placeholder="FZ" />
              </label>
              <label className="field compact-field">
                <span>Cilindraje</span>
                <input className="plain-input" type="number" min="1" value={motoForm.cilindraje} onChange={(event) => setMotoForm((current) => ({ ...current, cilindraje: event.target.value }))} placeholder="150" />
              </label>
              <label className="field compact-field">
                <span>Anio</span>
                <input className="plain-input" type="number" min="1950" value={motoForm.anio} onChange={(event) => setMotoForm((current) => ({ ...current, anio: event.target.value }))} placeholder="2024" />
              </label>
              <label className="field compact-field">
                <span>Color</span>
                <input className="plain-input" value={motoForm.color} onChange={(event) => setMotoForm((current) => ({ ...current, color: event.target.value }))} placeholder="Negro" />
              </label>
              <label className="field compact-field">
                <span>Kilometraje</span>
                <input className="plain-input" type="number" min="0" value={motoForm.kilometraje_actual} onChange={(event) => setMotoForm((current) => ({ ...current, kilometraje_actual: event.target.value }))} placeholder="0" />
              </label>
              {motoError ? <p className="form-error">{motoError}</p> : null}
              {motoSuccess ? <p className="form-success">{motoSuccess}</p> : null}
              <div className="catalog-form-actions">
                <button className="submit-button" type="submit">{editingMotoID ? "Guardar cambios" : "Crear motocicleta"}</button>
                {editingMotoID ? <button className="secondary-light-button" type="button" onClick={limpiarMotoForm}>Cancelar</button> : null}
              </div>
            </div>
          </form>
          <div className="catalog-card-list">
            <div className="catalog-search">
              <input className="search-input" value={busquedaMotos} onChange={(event) => setBusquedaMotos(event.target.value)} placeholder="Buscar moto" />
              <button className="secondary-light-button" type="button" onClick={cargarMotos}>Buscar</button>
            </div>
            <CatalogMotosTable motos={motos} editarMoto={editarMoto} />
          </div>
          </article>
        </section>
      ) : null}
    </AppShell>
  );
}

function CatalogItemsTable({ items, editarItem, desactivarItem }) {
  return (
    <div className="catalog-table" role="table" aria-label="Catalogo de items">
      <div className="catalog-row-head" role="row">
        <span>Item</span>
        <span>Tipo</span>
        <span>Estado</span>
        <span>Acciones</span>
      </div>
      {items.length ? items.map((item) => (
        <div className="catalog-row-item" role="row" key={item.id}>
          <div>
            <strong>{item.nombre}</strong>
            <span>{item.descripcion || "Sin descripcion"}</span>
          </div>
          <span>{tipoLabel(item.tipo_item)}</span>
          <span className={`state-pill ${item.activo ? "is-active" : "is-inactive"}`}>{item.activo ? "Activo" : "Inactivo"}</span>
          <div className="row-actions">
            <button className="secondary-light-button" type="button" onClick={() => editarItem(item)}>Editar</button>
            {item.activo ? <button className="remove-item-button" type="button" onClick={() => desactivarItem(item.id)}>Desactivar</button> : null}
          </div>
        </div>
      )) : (
        <div className="empty-state">
          <strong>Sin items registrados</strong>
          <span>Agrega el primer item del catalogo.</span>
        </div>
      )}
    </div>
  );
}

function CatalogClientesTable({ clientes, editarCliente, desactivarCliente }) {
  return (
    <div className="catalog-table" role="table" aria-label="Catalogo de clientes">
      <div className="catalog-row-head catalog-client-row" role="row">
        <span>Cliente</span>
        <span>Correo</span>
        <span>Telefono</span>
        <span>Acciones</span>
      </div>
      {clientes.length ? clientes.map((cliente) => (
        <div className="catalog-row-item catalog-client-row" role="row" key={cliente.id}>
          <div>
            <strong>{nombreCliente(cliente)}</strong>
            <span>{cliente.activo ? "Activo" : "Inactivo"}</span>
          </div>
          <span>{cliente.correo || "Sin correo"}</span>
          <span>{cliente.telefono || "Sin telefono"}</span>
          <div className="row-actions">
            <button className="secondary-light-button" type="button" onClick={() => editarCliente(cliente)}>Editar</button>
            {cliente.activo ? <button className="remove-item-button" type="button" onClick={() => desactivarCliente(cliente.id)}>Desactivar</button> : null}
          </div>
        </div>
      )) : (
        <div className="empty-state">
          <strong>Sin clientes registrados</strong>
          <span>Agrega el primer cliente del catalogo.</span>
        </div>
      )}
    </div>
  );
}

function CatalogMotosTable({ motos, editarMoto }) {
  return (
    <div className="catalog-table" role="table" aria-label="Catalogo de motocicletas">
      <div className="catalog-row-head catalog-moto-row" role="row">
        <span>Motocicleta</span>
        <span>Cliente</span>
        <span>Kilometraje</span>
        <span>Acciones</span>
      </div>
      {motos.length ? motos.map((moto) => (
        <div className="catalog-row-item catalog-moto-row" role="row" key={moto.id}>
          <div>
            <strong>{moto.placa} - {moto.marca} {moto.modelo}</strong>
            <span>{moto.cilindraje} cc {moto.anio ? `- ${moto.anio}` : ""} {moto.color ? `- ${moto.color}` : ""}</span>
          </div>
          <span>{moto.nombre_cliente || "Sin cliente"}</span>
          <span>{Number(moto.kilometraje_actual || 0).toLocaleString("es-CO")} km</span>
          <div className="row-actions">
            <button className="secondary-light-button" type="button" onClick={() => editarMoto(moto)}>Editar</button>
          </div>
        </div>
      )) : (
        <div className="empty-state">
          <strong>Sin motocicletas registradas</strong>
          <span>Agrega la primera motocicleta del catalogo.</span>
        </div>
      )}
    </div>
  );
}

function tipoLabel(value) {
  return tipos.find(([tipo]) => tipo === value)?.[1] || value;
}

function catalogoTitulo(value) {
  if (value === "items") return "Items Orden de Trabajo";
  if (value === "clientes") return "Clientes";
  if (value === "motos") return "Motocicletas";
  return "Catalogos del taller";
}

function nombreCliente(cliente) {
  return [cliente.nombres, cliente.apellidos].filter(Boolean).join(" ").trim() || "Cliente sin nombre";
}
