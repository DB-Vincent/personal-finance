import { Outlet, createRootRouteWithContext, Link } from "@tanstack/react-router";
import { Toaster } from "@/components/ui/sonner";

interface RouterContext {
  auth: {
    isAuthenticated: boolean;
  };
}

function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4 text-center px-4">
      <p className="text-6xl font-bold text-muted-foreground">404</p>
      <h1 className="text-2xl font-semibold">Page not found</h1>
      <p className="text-muted-foreground">The page you're looking for doesn't exist.</p>
      <Link
        to="/"
        className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
      >
        Go home
      </Link>
    </div>
  );
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: () => (
    <>
      <Outlet />
      <Toaster />
    </>
  ),
  notFoundComponent: NotFoundPage,
});
