import React, { FormEvent, useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import './styles.css';

type User = { id: string; email: string; full_name: string; role: 'ADMIN' | 'CLIENT' | string };
type Product = { id: string; name: string; price: number };
type BakePlan = { id: string; product_id: string; product_name?: string; plan_date: string; planned_quantity: number; available_quantity: number };
type OrderItem = { id: string; product_id: string; bake_plan_id: string; quantity: number; unit_price: number };
type Order = { id: string; client_id: string; store_name: string; status: string; items: OrderItem[]; created_at?: string };
type IngredientSection = { name: string; amount: string; unit: string };
type Ingredient = { id: string; name: string; unit: string; cost_per_unit: number; sections?: IngredientSection[] };
type Task = { id: string; title: string; status: string; due_date?: string; created_at?: string };

type ToastType = 'success' | 'error' | 'info';
type Toast = { type: ToastType; message: string } | null;
type Page = 'dashboard' | 'products-plans' | 'orders' | 'ingredients-recipes' | 'tasks' | 'client';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
const today = new Date().toISOString().slice(0, 10);
const PRODUCT_OPTIONS = ['Bread', 'Cake', 'Croissant', 'Bun', 'Cookie', 'Donut', 'Baguette', 'Cheesecake'];
const TASK_STATUSES = ['TODO', 'IN_PROGRESS', 'DONE', 'CANCELLED'];

function App() {
  const [token, setToken] = useState(localStorage.getItem('bakeplan_token') || '');
  const [user, setUser] = useState<User | null>(() => {
    const raw = localStorage.getItem('bakeplan_user');
    if (!raw) return null;
    try { return JSON.parse(raw); } catch { return null; }
  });
  const [page, setPage] = useState<Page>(() => {
    const raw = localStorage.getItem('bakeplan_user');
    try {
      const stored = raw ? JSON.parse(raw) : null;
      return stored?.role === 'CLIENT' ? 'client' : 'dashboard';
    } catch { return 'dashboard'; }
  });
  const [toast, setToast] = useState<Toast>(null);
  const [loading, setLoading] = useState(false);

  const [products, setProducts] = useState<Product[]>([]);
  const [bakePlans, setBakePlans] = useState<BakePlan[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [ingredients, setIngredients] = useState<Ingredient[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);

  const [planDate, setPlanDate] = useState(today);
  const [clientDate, setClientDate] = useState(today);

  const notify = (type: ToastType, message: string) => {
    setToast({ type, message });
    window.setTimeout(() => setToast(null), 3600);
  };

  const headers = useMemo(() => {
    const h: Record<string, string> = { 'Content-Type': 'application/json' };
    if (token) h.Authorization = `Bearer ${token}`;
    return h;
  }, [token]);

  async function request<T>(path: string, options: RequestInit = {}, silent = false): Promise<T> {
    if (!silent) setLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}${path}`, {
        ...options,
        headers: { ...headers, ...(options.headers || {}) }
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(data.error || data.message || `Request failed with status ${response.status}`);
      }
      return data as T;
    } catch (error) {
      const message = error instanceof Error ? cleanError(error.message) : 'Something went wrong';
      if (!silent) notify('error', message);
      throw error;
    } finally {
      if (!silent) setLoading(false);
    }
  }

  function saveSession(nextToken: string, nextUser: User) {
    localStorage.setItem('bakeplan_token', nextToken);
    localStorage.setItem('bakeplan_user', JSON.stringify(nextUser));
    setToken(nextToken);
    setUser(nextUser);
    setPage(nextUser.role === 'CLIENT' ? 'client' : 'dashboard');
  }

  function logout() {
    localStorage.removeItem('bakeplan_token');
    localStorage.removeItem('bakeplan_user');
    setToken('');
    setUser(null);
    setPage('dashboard');
    notify('info', 'You have logged out.');
  }

  async function login(email: string, password: string) {
    const data = await request<{ user: User; access_token: string }>('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) });
    saveSession(data.access_token, data.user);
    notify('success', `Welcome back, ${data.user.full_name}.`);
  }

  async function register(fullName: string, email: string, password: string, role: string) {
    const data = await request<{ user: User; token: string }>('/auth/register', { method: 'POST', body: JSON.stringify({ full_name: fullName, email, password, role }) });
    saveSession(data.token, data.user);
    notify('success', `Account created for ${data.user.full_name}.`);
  }

  async function loadProducts(silent = true) {
    const data = await request<{ products: Product[] }>('/products', {}, silent);
    setProducts(data.products || []);
  }

  async function createProduct(name: string, price: number) {
    if (!name) return notify('error', 'Please choose a product.');
    if (!price || price <= 0) return notify('error', 'Price must be greater than zero.');
    await request('/admin/products', { method: 'POST', body: JSON.stringify({ name, price }) });
    notify('success', `${name} was added to the product catalog.`);
    await loadProducts(false);
  }

  async function loadBakePlans(date = planDate, silent = true) {
    const data = await request<{ bake_plans: BakePlan[] }>(`/bake-plans?date=${encodeURIComponent(date)}`, {}, silent);
    setBakePlans(data.bake_plans || []);
  }

  async function loadAvailableProducts(date = clientDate, silent = true) {
    const data = await request<{ bake_plans: BakePlan[] }>(`/available-products?date=${encodeURIComponent(date)}`, {}, silent);
    setBakePlans(data.bake_plans || []);
  }

  async function createBakePlan(productId: string, planDateValue: string, plannedQuantity: number) {
    if (!productId) return notify('error', 'Please select a product first.');
    if (!planDateValue) return notify('error', 'Please choose a baking date.');
    if (!plannedQuantity || plannedQuantity <= 0) return notify('error', 'Planned quantity must be greater than zero.');
    await request('/admin/bake-plans', { method: 'POST', body: JSON.stringify({ product_id: productId, plan_date: planDateValue, planned_quantity: plannedQuantity }) });
    notify('success', 'Baking plan was created.');
    await loadBakePlans(planDateValue, false);
  }

  async function loadOrders(silent = true) {
    const path = user?.role === 'ADMIN' ? '/admin/orders' : '/orders/my';
    const data = await request<{ orders: Order[] }>(path, {}, silent);
    setOrders(data.orders || []);
  }

  async function updateOrderStatus(orderId: string, status: string) {
    await request('/admin/orders/status', { method: 'PATCH', body: JSON.stringify({ id: orderId, status }) });
    notify('success', `Order marked as ${status}.`);
    await loadOrders(false);
  }

  async function createOrder(storeName: string, bakePlanId: string, quantity: number) {
    const selectedPlan = bakePlans.find((plan) => plan.id === bakePlanId);
    if (!storeName.trim()) return notify('error', 'Store name is required.');
    if (!bakePlanId || !selectedPlan) return notify('error', 'Please select an available baking plan.');
    if (!quantity || quantity <= 0) return notify('error', 'Quantity must be greater than zero.');
    if (quantity > selectedPlan.available_quantity) {
      return notify('error', `Not enough quantity. Only ${selectedPlan.available_quantity} items are available.`);
    }
    await request('/orders', { method: 'POST', body: JSON.stringify({ store_name: storeName, bake_plan_id: bakePlanId, quantity }) });
    notify('success', 'Order was created successfully.');
    await Promise.all([loadOrders(false), user?.role === 'CLIENT' ? loadAvailableProducts(clientDate, false) : loadBakePlans(planDate, false)]);
  }

  async function loadIngredients(silent = true) {
    const data = await request<{ ingredients: Ingredient[] }>('/admin/ingredients', {}, silent);
    setIngredients(data.ingredients || []);
  }

  async function createIngredient(payload: { name: string; unit: string; cost_per_unit: number; sections: IngredientSection[] }) {
    if (!payload.name.trim()) return notify('error', 'Name is required.');
    if (!payload.unit.trim()) return notify('error', 'Unit is required.');
    if (payload.cost_per_unit < 0) return notify('error', 'Cost per unit cannot be negative.');
    await request('/admin/ingredients', { method: 'POST', body: JSON.stringify(payload) });
    notify('success', `${payload.name} was saved and added to Products & Plans.`);
    await Promise.all([loadIngredients(false), loadProducts(false)]);
  }

  async function loadTasks(silent = true) {
    const data = await request<{ tasks: Task[] }>('/admin/tasks', {}, silent);
    setTasks(data.tasks || []);
  }

  async function createTask(title: string, dueDate: string) {
    if (!title.trim()) return notify('error', 'Task title is required.');
    await request('/admin/tasks', { method: 'POST', body: JSON.stringify({ title, due_date: dueDate }) });
    notify('success', 'Task was created.');
    await loadTasks(false);
  }

  async function updateTaskStatus(task: Task, status: string) {
    await request('/admin/tasks/status', { method: 'PATCH', body: JSON.stringify({ id: task.id, title: task.title, status, due_date: task.due_date || '' }) });
    notify('success', 'Task status updated.');
    await loadTasks(false);
  }

  async function deleteTask(id: string) {
    await request(`/admin/tasks/${id}`, { method: 'DELETE' });
    notify('success', 'Task was removed.');
    await loadTasks(false);
  }

  useEffect(() => {
    if (!token) return;
    request<{ user: User }>('/auth/me', {}, true)
      .then((data) => {
        setUser(data.user);
        localStorage.setItem('bakeplan_user', JSON.stringify(data.user));
        if (data.user.role === 'CLIENT') setPage('client');
      })
      .catch(() => logout());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  useEffect(() => {
    if (!user) return;
    loadProducts(true).catch(() => undefined);
    if (user.role === 'ADMIN') {
      loadBakePlans(planDate, true).catch(() => undefined);
      loadOrders(true).catch(() => undefined);
      loadIngredients(true).catch(() => undefined);
      loadTasks(true).catch(() => undefined);
    } else {
      loadAvailableProducts(clientDate, true).catch(() => undefined);
      loadOrders(true).catch(() => undefined);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  if (!token || !user) {
    return <AuthPage onLogin={login} onRegister={register} loading={loading} toast={toast} />;
  }

  const isAdmin = user.role === 'ADMIN';
  const nav = isAdmin
    ? [
        ['dashboard', 'Dashboard'],
        ['products-plans', 'Products & Plans'],
        ['orders', 'Orders'],
        ['ingredients-recipes', 'Ingredients'],
        ['tasks', 'Tasks']
      ] as Array<[Page, string]>
    : [['client', 'Store page']] as Array<[Page, string]>;

  const totalRevenue = orders.reduce((sum, order) => sum + (order.items || []).reduce((itemSum, item) => itemSum + item.quantity * item.unit_price, 0), 0);
  const plannedTotal = bakePlans.reduce((sum, p) => sum + p.planned_quantity, 0);
  const availableTotal = bakePlans.reduce((sum, p) => sum + p.available_quantity, 0);
  const orderedTotal = Math.max(0, plannedTotal - availableTotal);

  return (
    <main className="app-shell">
      <ToastView toast={toast} />
      <aside className="sidebar">
        <div className="brand-block">
          <div className="brand-mark">B</div>
          <div>
            <h1>BakePlan</h1>
            <p>Bakery management</p>
          </div>
        </div>
        <nav className="side-nav">
          {nav.map(([id, label]) => (
            <button key={id} className={page === id ? 'active' : ''} onClick={() => setPage(id)}>{label}</button>
          ))}
        </nav>
        <div className="user-card">
          <span className="role-pill">{user.role === 'ADMIN' ? 'Administrator' : 'Store client'}</span>
          <strong>{user.full_name}</strong>
          <small>{user.email}</small>
          <button className="ghost" onClick={logout}>Log out</button>
        </div>
      </aside>

      <section className="workspace">
        <header className="topbar">
          <div>
            <span className="eyebrow">Production and returns</span>
            <h2>{nav.find(([id]) => id === page)?.[1]}</h2>
          </div>
          {loading && <div className="saving-dot">Saving...</div>}
        </header>

        {isAdmin && page === 'dashboard' && (
          <AdminDashboard
            products={products}
            bakePlans={bakePlans}
            orders={orders}
            tasks={tasks}
            totalRevenue={totalRevenue}
            plannedTotal={plannedTotal}
            orderedTotal={orderedTotal}
            availableTotal={availableTotal}
            onReload={() => Promise.all([loadProducts(false), loadBakePlans(planDate, false), loadOrders(false), loadTasks(false)])}
          />
        )}

        {isAdmin && page === 'products-plans' && (
          <ProductsPlansPage
            products={products}
            bakePlans={bakePlans}
            planDate={planDate}
            setPlanDate={setPlanDate}
            createProduct={createProduct}
            createBakePlan={createBakePlan}
            loadBakePlans={loadBakePlans}
          />
        )}

        {isAdmin && page === 'orders' && (
          <OrdersPage
            bakePlans={bakePlans}
            orders={orders}
            createOrder={createOrder}
            updateOrderStatus={updateOrderStatus}
            loadOrders={() => loadOrders(false)}
          />
        )}

        {isAdmin && page === 'ingredients-recipes' && (
          <IngredientsPage ingredients={ingredients} createIngredient={createIngredient} loadIngredients={() => loadIngredients(false)} />
        )}

        {isAdmin && page === 'tasks' && (
          <TasksPage tasks={tasks} createTask={createTask} updateTaskStatus={updateTaskStatus} deleteTask={deleteTask} loadTasks={() => loadTasks(false)} />
        )}

        {!isAdmin && page === 'client' && (
          <ClientPage
            user={user}
            date={clientDate}
            setDate={setClientDate}
            bakePlans={bakePlans}
            orders={orders}
            loadAvailableProducts={loadAvailableProducts}
            createOrder={createOrder}
            loadOrders={() => loadOrders(false)}
          />
        )}
      </section>
    </main>
  );
}

function AuthPage({ onLogin, onRegister, loading, toast }: { onLogin: (email: string, password: string) => Promise<void>; onRegister: (name: string, email: string, password: string, role: string) => Promise<void>; loading: boolean; toast: Toast }) {
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [name, setName] = useState('Admin');
  const [email, setEmail] = useState('admin@bakeplan.kz');
  const [password, setPassword] = useState('123456');
  const [role, setRole] = useState('ADMIN');

  async function submit(e: FormEvent) {
    e.preventDefault();
    if (mode === 'login') await onLogin(email, password);
    else await onRegister(name, email, password, role);
  }

  return (
    <main className="auth-page">
      <ToastView toast={toast} />
      <section className="auth-hero">
        <div className="auth-logo">🥐</div>
        <h1>Fresh bakery planning, clean production control.</h1>
        <p>Manage products, baking plans, orders, tasks, ingredients, and store demand in one calm workspace.</p>
        <div className="hero-stats">
          <span><b>3</b> microservices</span>
          <span><b>1</b> gateway</span>
          <span><b>Real</b> bakery flow</span>
        </div>
      </section>
      <form className="auth-card" onSubmit={submit}>
        <div className="auth-switch">
          <button type="button" className={mode === 'login' ? 'active' : ''} onClick={() => setMode('login')}>Login</button>
          <button type="button" className={mode === 'register' ? 'active' : ''} onClick={() => setMode('register')}>Register</button>
        </div>
        <h2>{mode === 'login' ? 'Welcome back' : 'Create an account'}</h2>
        {mode === 'register' && <Field label="Full name"><input value={name} onChange={(e) => setName(e.target.value)} placeholder="Your name" /></Field>}
        <Field label="Email"><input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="email@example.com" /></Field>
        <Field label="Password"><input value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Password" type="password" /></Field>
        {mode === 'register' && (
          <Field label="Role">
            <select value={role} onChange={(e) => setRole(e.target.value)}>
              <option value="ADMIN">Admin</option>
              <option value="CLIENT">Client / Store</option>
            </select>
          </Field>
        )}
        <button className="primary full" disabled={loading}>{loading ? 'Please wait...' : mode === 'login' ? 'Login' : 'Create account'}</button>
      </form>
    </main>
  );
}

function AdminDashboard({ products, bakePlans, orders, tasks, totalRevenue, plannedTotal, orderedTotal, availableTotal, onReload }: { products: Product[]; bakePlans: BakePlan[]; orders: Order[]; tasks: Task[]; totalRevenue: number; plannedTotal: number; orderedTotal: number; availableTotal: number; onReload: () => void }) {
  const activeTasks = tasks.filter((t) => t.status !== 'DONE' && t.status !== 'CANCELLED').length;
  return (
    <div className="page-stack">
      <section className="welcome-card">
        <div>
          <span className="eyebrow">Today overview</span>
          <h3>Bakery production dashboard</h3>
          <p>Statistics are combined here, so admins can quickly see production, orders, leftovers, and income.</p>
        </div>
        <button className="secondary" onClick={onReload}>Reload data</button>
      </section>
      <div className="metrics-grid">
        <Metric title="Products" value={products.length} note="catalog items" />
        <Metric title="Planned" value={plannedTotal} note="pieces" />
        <Metric title="Ordered" value={orderedTotal} note="pieces" />
        <Metric title="Left" value={availableTotal} note="pieces" />
        <Metric title="Orders" value={orders.length} note="total" />
        <Metric title="Open tasks" value={activeTasks} note="need attention" />
        <Metric title="Revenue" value={`${money(totalRevenue)}`} note="from orders" wide />
      </div>
      <section className="panel">
        <PanelHeader title="Production statistics" subtitle="Planned quantity, ordered quantity, leftovers, and estimated revenue by baking plan." />
        <DataTable
          headers={['Product', 'Date', 'Planned', 'Ordered', 'Left', 'Sell-through']}
          rows={bakePlans.map((plan) => {
            const ordered = plan.planned_quantity - plan.available_quantity;
            const rate = plan.planned_quantity ? `${Math.round((ordered / plan.planned_quantity) * 100)}%` : '0%';
            return [plan.product_name || plan.product_id, plan.plan_date, plan.planned_quantity, ordered, plan.available_quantity, <Badge key={plan.id} text={rate} tone={ordered > 0 ? 'good' : 'neutral'} />];
          })}
        />
      </section>
    </div>
  );
}

function ProductsPlansPage({ products, bakePlans, planDate, setPlanDate, createProduct, createBakePlan, loadBakePlans }: { products: Product[]; bakePlans: BakePlan[]; planDate: string; setPlanDate: (v: string) => void; createProduct: (name: string, price: number) => Promise<void>; createBakePlan: (productId: string, date: string, qty: number) => Promise<void>; loadBakePlans: (date?: string, silent?: boolean) => Promise<void> }) {
  const [productName, setProductName] = useState(PRODUCT_OPTIONS[0]);
  const [price, setPrice] = useState('200');
  const [selectedProduct, setSelectedProduct] = useState('');
  const [planQty, setPlanQty] = useState('300');

  useEffect(() => {
    if (!selectedProduct && products.length > 0) setSelectedProduct(products[0].id);
  }, [products, selectedProduct]);

  return (
    <div className="page-stack">
      <div className="two-columns">
        <form className="panel" onSubmit={(e) => { e.preventDefault(); createProduct(productName, Number(price)); }}>
          <PanelHeader title="Product catalog" subtitle="Choose a bakery item and add its selling price." />
          <div className="form-grid compact">
            <Field label="Product"><select value={productName} onChange={(e) => setProductName(e.target.value)}>{PRODUCT_OPTIONS.map((p) => <option key={p}>{p}</option>)}</select></Field>
            <Field label="Price"><input value={price} onChange={(e) => setPrice(e.target.value)} type="number" min="1" /></Field>
          </div>
          <button className="primary">Add product</button>
        </form>

        <form className="panel" onSubmit={(e) => { e.preventDefault(); createBakePlan(selectedProduct, planDate, Number(planQty)); }}>
          <PanelHeader title="Baking plan" subtitle="Create production only from already saved products." />
          <div className="form-grid compact">
            <Field label="Product"><select value={selectedProduct} onChange={(e) => setSelectedProduct(e.target.value)}><option value="">Select product</option>{products.map((p) => <option key={p.id} value={p.id}>{p.name} — {money(p.price)}</option>)}</select></Field>
            <Field label="Date"><input value={planDate} onChange={(e) => setPlanDate(e.target.value)} type="date" /></Field>
            <Field label="Planned quantity"><input value={planQty} onChange={(e) => setPlanQty(e.target.value)} type="number" min="1" /></Field>
          </div>
          <button className="primary">Create plan</button>
        </form>
      </div>

      <section className="panel">
        <PanelHeader title="Products" subtitle="Saved products that can be used in baking plans." />
        <DataTable headers={['Name', 'Price']} rows={products.map((p) => [p.name, money(p.price)])} />
      </section>

      <section className="panel">
        <div className="panel-heading split">
          <div><h3>Baking plans</h3><p>Filter plans by production date.</p></div>
          <div className="inline-control"><input value={planDate} onChange={(e) => setPlanDate(e.target.value)} type="date" /><button className="secondary" onClick={() => loadBakePlans(planDate, false)}>Load</button></div>
        </div>
        <DataTable headers={['Product', 'Date', 'Planned', 'Available']} rows={bakePlans.map((p) => [p.product_name || p.product_id, p.plan_date, p.planned_quantity, <Badge key={p.id} text={`${p.available_quantity} left`} tone={p.available_quantity > 0 ? 'good' : 'danger'} />])} />
      </section>
    </div>
  );
}

function OrdersPage({ bakePlans, orders, createOrder, updateOrderStatus, loadOrders }: { bakePlans: BakePlan[]; orders: Order[]; createOrder: (storeName: string, bakePlanId: string, quantity: number) => Promise<void>; updateOrderStatus: (orderId: string, status: string) => Promise<void>; loadOrders: () => void }) {
  const [storeName, setStoreName] = useState('Store A');
  const [bakePlanId, setBakePlanId] = useState('');
  const [quantity, setQuantity] = useState('30');
  useEffect(() => { if (!bakePlanId && bakePlans.length > 0) setBakePlanId(bakePlans[0].id); }, [bakePlans, bakePlanId]);
  return (
    <div className="page-stack">
      <form className="panel horizontal-form" onSubmit={(e) => { e.preventDefault(); createOrder(storeName, bakePlanId, Number(quantity)); }}>
        <PanelHeader title="Create order" subtitle="The system checks available quantity before saving." />
        <Field label="Store"><input value={storeName} onChange={(e) => setStoreName(e.target.value)} /></Field>
        <Field label="Baking plan"><select value={bakePlanId} onChange={(e) => setBakePlanId(e.target.value)}><option value="">Select plan</option>{bakePlans.map((p) => <option key={p.id} value={p.id}>{p.product_name || p.product_id} — {p.plan_date} — {p.available_quantity} left</option>)}</select></Field>
        <Field label="Quantity"><input value={quantity} onChange={(e) => setQuantity(e.target.value)} type="number" min="1" /></Field>
        <button className="primary align-bottom">Create</button>
      </form>
      <section className="panel">
        <div className="panel-heading split"><div><h3>Orders</h3><p>Vertical order cards for easier reading.</p></div><button className="secondary" onClick={loadOrders}>Refresh</button></div>
        <OrderList orders={orders} onStatusChange={updateOrderStatus} />
      </section>
    </div>
  );
}

function IngredientsPage({ ingredients, createIngredient, loadIngredients }: { ingredients: Ingredient[]; createIngredient: (payload: { name: string; unit: string; cost_per_unit: number; sections: IngredientSection[] }) => Promise<void>; loadIngredients: () => void }) {
  const [name, setName] = useState('Bread');
  const [unit, setUnit] = useState('1 kg');
  const [cost, setCost] = useState('200');
  const [sections, setSections] = useState<IngredientSection[]>([{ name: 'Salt', amount: '10', unit: 'gr' }, { name: 'Flour', amount: '600', unit: 'gr' }]);

  function updateSection(index: number, key: keyof IngredientSection, value: string) {
    setSections((prev) => prev.map((item, i) => i === index ? { ...item, [key]: value } : item));
  }

  return (
    <div className="page-stack">
      <form className="panel" onSubmit={(e) => { e.preventDefault(); createIngredient({ name, unit, cost_per_unit: Number(cost), sections: sections.filter((s) => s.name.trim()) }); }}>
        <PanelHeader title="Create ingredient or recipe base" subtitle="Default fields are Name, Unit, and Cost per unit. Add composition rows when needed." />
        <div className="form-grid compact">
          <Field label="Name"><input value={name} onChange={(e) => setName(e.target.value)} placeholder="Bread" /></Field>
          <Field label="Unit"><input value={unit} onChange={(e) => setUnit(e.target.value)} placeholder="1 kg" /></Field>
          <Field label="Cost per unit"><input value={cost} onChange={(e) => setCost(e.target.value)} type="number" min="0" placeholder="200" /></Field>
        </div>
        <div className="dynamic-list">
          <div className="dynamic-title">Additional sections</div>
          {sections.map((section, index) => (
            <div className="recipe-row" key={index}>
              <input value={section.name} onChange={(e) => updateSection(index, 'name', e.target.value)} placeholder="Salt" />
              <input value={section.amount} onChange={(e) => updateSection(index, 'amount', e.target.value)} placeholder="10" />
              <input value={section.unit} onChange={(e) => updateSection(index, 'unit', e.target.value)} placeholder="gr" />
              <button type="button" className="ghost danger-text" onClick={() => setSections((prev) => prev.filter((_, i) => i !== index))}>Remove</button>
            </div>
          ))}
          <button type="button" className="secondary" onClick={() => setSections((prev) => [...prev, { name: '', amount: '', unit: '' }])}>+ Add section</button>
        </div>
        <button className="primary">Save</button>
      </form>

      <section className="panel">
        <div className="panel-heading split"><div><h3>Saved ingredients</h3><p>Ingredient and recipe bases.</p></div><button className="secondary" onClick={loadIngredients}>Refresh</button></div>
        <div className="ingredient-grid">
          {ingredients.length === 0 && <Empty message="No ingredients yet." />}
          {ingredients.map((ingredient) => <IngredientCard key={ingredient.id} ingredient={ingredient} />)}
        </div>
      </section>
    </div>
  );
}

function TasksPage({ tasks, createTask, updateTaskStatus, deleteTask, loadTasks }: { tasks: Task[]; createTask: (title: string, dueDate: string) => Promise<void>; updateTaskStatus: (task: Task, status: string) => Promise<void>; deleteTask: (id: string) => Promise<void>; loadTasks: () => void }) {
  const [title, setTitle] = useState('Prepare dough');
  const [dueDate, setDueDate] = useState(today);
  return (
    <div className="page-stack">
      <form className="panel horizontal-form" onSubmit={(e) => { e.preventDefault(); createTask(title, dueDate); setTitle(''); }}>
        <PanelHeader title="Create task" subtitle="Creation is simple. Status is managed from the task list below." />
        <Field label="Task title"><input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Buy flour" /></Field>
        <Field label="Due date"><input value={dueDate} onChange={(e) => setDueDate(e.target.value)} type="date" /></Field>
        <button className="primary align-bottom">Create task</button>
      </form>
      <section className="panel">
        <div className="panel-heading split"><div><h3>Tasks</h3><p>Change TODO, IN_PROGRESS, DONE, or CANCELLED. Remove finished or wrong tasks.</p></div><button className="secondary" onClick={loadTasks}>Refresh</button></div>
        <div className="task-list">
          {tasks.length === 0 && <Empty message="No tasks yet." />}
          {tasks.map((task) => (
            <article className="task-card" key={task.id}>
              <div><strong>{task.title}</strong><small>{task.due_date ? `Due ${task.due_date}` : 'No due date'}</small></div>
              <select value={task.status} onChange={(e) => updateTaskStatus(task, e.target.value)}>{TASK_STATUSES.map((s) => <option key={s}>{s}</option>)}</select>
              <button className="ghost danger-text" onClick={() => deleteTask(task.id)}>Delete</button>
            </article>
          ))}
        </div>
      </section>
    </div>
  );
}

function ClientPage({ user, date, setDate, bakePlans, orders, loadAvailableProducts, createOrder, loadOrders }: { user: User; date: string; setDate: (v: string) => void; bakePlans: BakePlan[]; orders: Order[]; loadAvailableProducts: (date?: string, silent?: boolean) => Promise<void>; createOrder: (storeName: string, bakePlanId: string, quantity: number) => Promise<void>; loadOrders: () => void }) {
  const [storeName, setStoreName] = useState(user.full_name || 'Store A');
  const [bakePlanId, setBakePlanId] = useState('');
  const [quantity, setQuantity] = useState('10');
  useEffect(() => { if (!bakePlanId && bakePlans.length > 0) setBakePlanId(bakePlans[0].id); }, [bakePlans, bakePlanId]);
  return (
    <div className="page-stack">
      <section className="welcome-card client">
        <div><span className="eyebrow">Store workspace</span><h3>Order fresh bakery products</h3><p>Choose from available baking plans. Your order automatically decreases available quantity.</p></div>
        <div className="inline-control"><input value={date} onChange={(e) => setDate(e.target.value)} type="date" /><button className="secondary" onClick={() => loadAvailableProducts(date, false)}>Load products</button></div>
      </section>
      <section className="panel">
        <PanelHeader title="Available products" subtitle="Only products with remaining quantity are shown." />
        <DataTable headers={['Product', 'Date', 'Available']} rows={bakePlans.map((p) => [p.product_name || p.product_id, p.plan_date, <Badge key={p.id} text={`${p.available_quantity} pcs`} tone="good" />])} />
      </section>
      <form className="panel horizontal-form" onSubmit={(e) => { e.preventDefault(); createOrder(storeName, bakePlanId, Number(quantity)); }}>
        <PanelHeader title="Create order" subtitle="The form checks quantity before sending." />
        <Field label="Store"><input value={storeName} onChange={(e) => setStoreName(e.target.value)} /></Field>
        <Field label="Product"><select value={bakePlanId} onChange={(e) => setBakePlanId(e.target.value)}><option value="">Select available product</option>{bakePlans.map((p) => <option key={p.id} value={p.id}>{p.product_name || p.product_id} — {p.available_quantity} left</option>)}</select></Field>
        <Field label="Quantity"><input value={quantity} onChange={(e) => setQuantity(e.target.value)} type="number" min="1" /></Field>
        <button className="primary align-bottom">Order</button>
      </form>
      <section className="panel">
        <div className="panel-heading split"><div><h3>My orders</h3><p>Your store order history.</p></div><button className="secondary" onClick={loadOrders}>Refresh</button></div>
        <OrderList orders={orders} />
      </section>
    </div>
  );
}

function Metric({ title, value, note, wide }: { title: string; value: string | number; note: string; wide?: boolean }) {
  return <section className={`metric-card ${wide ? 'wide' : ''}`}><span>{title}</span><strong>{value}</strong><small>{note}</small></section>;
}

function PanelHeader({ title, subtitle }: { title: string; subtitle: string }) {
  return <div className="panel-heading"><h3>{title}</h3><p>{subtitle}</p></div>;
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return <label className="field"><span>{label}</span>{children}</label>;
}

function DataTable({ headers, rows }: { headers: string[]; rows: Array<Array<React.ReactNode>> }) {
  return <div className="table-wrap"><table><thead><tr>{headers.map((h) => <th key={h}>{h}</th>)}</tr></thead><tbody>{rows.length === 0 ? <tr><td colSpan={headers.length}><Empty message="No data yet." /></td></tr> : rows.map((row, index) => <tr key={index}>{row.map((cell, i) => <td key={i}>{cell}</td>)}</tr>)}</tbody></table></div>;
}

function OrderList({ orders, onStatusChange }: { orders: Order[]; onStatusChange?: (orderId: string, status: string) => Promise<void> }) {
  if (orders.length === 0) return <Empty message="No orders yet." />;
  return <div className="order-list">{orders.map((order) => {
    const pieces = (order.items || []).reduce((sum, item) => sum + item.quantity, 0);
    const revenue = (order.items || []).reduce((sum, item) => sum + item.quantity * item.unit_price, 0);
    const tone = order.status === 'CANCELLED' ? 'danger' : order.status === 'CONFIRMED' || order.status === 'DELIVERED' ? 'good' : 'neutral';
    return (
      <article className="order-card" key={order.id}>
        <div><span className="muted-label">Store</span><strong>{order.store_name}</strong><small>{order.created_at ? new Date(order.created_at).toLocaleString() : 'Recently created'}</small></div>
        <div><span className="muted-label">Quantity</span><strong>{pieces} pcs</strong><small>{money(revenue)}</small></div>
        <Badge text={order.status} tone={tone} />
        {onStatusChange && (
          <div className="order-actions">
            {order.status === 'PENDING' && <button className="secondary" onClick={() => onStatusChange(order.id, 'CONFIRMED')}>Accept</button>}
            {order.status !== 'DELIVERED' && order.status !== 'CANCELLED' && <button className="secondary" onClick={() => onStatusChange(order.id, 'DELIVERED')}>Delivered</button>}
            {order.status !== 'CANCELLED' && <button className="ghost danger-text" onClick={() => onStatusChange(order.id, 'CANCELLED')}>Cancel</button>}
          </div>
        )}
      </article>
    );
  })}</div>;
}

function IngredientCard({ ingredient }: { ingredient: Ingredient }) {
  return <article className="ingredient-card"><div><span className="muted-label">Name</span><strong>{ingredient.name}</strong></div><div><span className="muted-label">Unit</span><b>{ingredient.unit}</b></div><div><span className="muted-label">Cost</span><b>{money(ingredient.cost_per_unit)}</b></div>{ingredient.sections && ingredient.sections.length > 0 && <div className="section-chips">{ingredient.sections.map((s, i) => <span key={i}>{s.name} · {s.amount} {s.unit}</span>)}</div>}</article>;
}

function Badge({ text, tone }: { text: string; tone: 'good' | 'danger' | 'neutral' }) {
  return <span className={`badge ${tone}`}>{text}</span>;
}

function ToastView({ toast }: { toast: Toast }) {
  return toast ? <div className={`toast ${toast.type}`}>{toast.message}</div> : null;
}

function Empty({ message }: { message: string }) {
  return <div className="empty">{message}</div>;
}

function money(value: number) {
  return `${new Intl.NumberFormat('en-US').format(Math.round(value || 0))} tg`;
}

function cleanError(message: string) {
  return message.replace(/^rpc error: code = [A-Za-z]+ desc = /, '').replace(/^rpc error: /, '');
}

createRoot(document.getElementById('root')!).render(<App />);
