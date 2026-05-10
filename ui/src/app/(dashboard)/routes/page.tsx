import { fetchRoutes } from '@/components/routes/routes-data';
import { RoutesClient } from '@/components/routes/routes-client';

export default async function RoutesPage() {
  const { routes, unavailable } = await fetchRoutes();
  return <RoutesClient routes={routes} unavailable={unavailable} />;
}
