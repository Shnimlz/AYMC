import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const routes: RouteRecordRaw[] = [
  {
    path: "/login",
    name: "Login",
    component: () => import("@/views/Login.vue"),
    meta: { requiresAuth: false, title: "Iniciar Sesión" },
  },
  {
    path: "/register",
    name: "Register",
    component: () => import("@/views/Register.vue"),
    meta: { requiresAuth: false, title: "Registrarse" },
  },
  {
    path: "/",
    component: () => import("@/layouts/MainLayout.vue"),
    meta: { requiresAuth: true },
    children: [
      {
        path: "",
        redirect: "/dashboard",
      },
      {
        path: "/dashboard",
        name: "Dashboard",
        component: () => import("@/views/Dashboard.vue"),
        meta: { title: "Panel de Control" },
      },
      {
        path: "/servers",
        name: "Servers",
        component: () => import("@/views/Servers/List.vue"),
        meta: { title: "Servidores" },
      },
      {
        path: "/servers/create",
        name: "ServerCreate",
        component: () => import("@/views/Servers/Create.vue"),
        meta: { title: "Crear Servidor" },
      },
      {
        path: "/servers/:id",
        name: "ServerDetail",
        component: () => import("@/views/Servers/Detail.vue"),
        meta: { title: "Detalle del Servidor" },
      },
      {
        path: "/marketplace",
        name: "Marketplace",
        component: () => import("@/views/Marketplace/Search.vue"),
        meta: { title: "Marketplace" },
      },
      {
        path: "/marketplace/:source/:id",
        name: "MarketplaceDetail",
        component: () => import("@/views/Marketplace/Detail.vue"),
        meta: { title: "Detalle del Plugin" },
      },
      {
        path: "/marketplace/installed",
        name: "MarketplaceInstalled",
        component: () => import("@/views/Marketplace/Installed.vue"),
        meta: { title: "Plugins Instalados" },
      },
      {
        path: "/backups",
        name: "Backups",
        component: () => import("@/views/Backups/List.vue"),
        meta: { title: "Respaldos" },
      },
      {
        path: "/backups/config",
        name: "BackupsConfig",
        component: () => import("@/views/Backups/Config.vue"),
        meta: { title: "Configuración de Respaldos" },
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// Navigation guard
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore();
  const requiresAuth = to.matched.some(
    (record) => record.meta.requiresAuth !== false
  );

  // Actualizar título de la página
  if (to.meta.title) {
    document.title = `${to.meta.title} - AYMC`;
  }

  // Si la ruta requiere autenticación y no está autenticado
  if (requiresAuth && !authStore.isAuthenticated) {
    next({ name: "Login", query: { redirect: to.fullPath } });
  }
  // Si está autenticado y va a login/register, redirigir al dashboard
  else if (
    authStore.isAuthenticated &&
    (to.name === "Login" || to.name === "Register")
  ) {
    next({ name: "Dashboard" });
  }
  // Continuar normalmente
  else {
    next();
  }
});

export default router;
