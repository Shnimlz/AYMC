// Augmentación de tipos para Vue Router
import "vue-router";

declare module "vue-router" {
  interface RouteMeta {
    requiresAuth?: boolean;
    requiresSSH?: boolean;
    requiresBackend?: boolean;
    title?: string;
  }
}
