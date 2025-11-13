import { createRouter, createWebHistory } from "vue-router";
import type { RouteRecordRaw } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { invoke } from "@tauri-apps/api/core";

const routes: RouteRecordRaw[] = [
  // Rutas de Onboarding (primera vez)
  {
    path: "/welcome",
    name: "Welcome",
    component: () => import("@/views/Onboarding/Welcome.vue"),
    meta: { requiresAuth: false, title: "Bienvenida" },
  },
  {
    path: "/ssh-setup",
    name: "SSHSetup",
    component: () => import("@/views/Onboarding/SSHSetup.vue"),
    meta: { requiresAuth: false, title: "Configuración SSH" },
  },
  {
    path: "/detection",
    name: "Detection",
    component: () => import("@/views/Onboarding/Detection.vue"),
    meta: {
      requiresAuth: false,
      requiresSSH: true,
      title: "Detección de Servicios",
    },
  },
  {
    path: "/installer",
    name: "Installer",
    component: () => import("@/views/Onboarding/Installer.vue"),
    meta: {
      requiresAuth: false,
      requiresSSH: true,
      title: "Instalación de AYMC",
    },
  },
  // Rutas de autenticación existentes
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
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore();
  const requiresAuth = to.matched.some(
    (record) => record.meta.requiresAuth !== false
  );

  // Actualizar título de la página
  if (to.meta.title) {
    document.title = `${to.meta.title} - AYMC`;
  }

  // Verificar si la ruta requiere conexión SSH
  if (to.meta.requiresSSH) {
    try {
      const isConnected = await invoke<boolean>("ssh_is_connected");

      if (!isConnected) {
        console.warn("SSH no conectado, redirigiendo a setup");
        next({ name: "SSHSetup" });
        return;
      }
    } catch (error) {
      console.error("Error verificando SSH:", error);
      next({ name: "SSHSetup" });
      return;
    }
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
