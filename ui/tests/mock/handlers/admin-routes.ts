import { Hono } from 'hono';
import { routes } from '../../fixtures/routes';

const app = new Hono();

app.get('/', (c) => c.json({ data: routes }));

app.post('/', (c) =>
  c.json({ error: { message: 'route mutations not available', type: 'not_implemented_error' } }, 501)
);

app.patch('/:id', (c) =>
  c.json({ error: { message: 'route mutations not available', type: 'not_implemented_error' } }, 501)
);

export default app;
