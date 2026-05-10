import { Hono } from "hono";
import { getStore } from "../state";

const app = new Hono();

app.get("/", (c) => c.json({ data: getStore().teams }));

export default app;
