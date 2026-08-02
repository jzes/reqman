import { sleep } from "bun";
import { Hono } from "hono";
import { ContentfulStatusCode } from "hono/utils/http-status";

const app = new Hono();

app.get("/", (c) => {
	return c.json({ message: "hello from hono" });
});

app.get("/get/:status", async (ctx) => {
	const status = Number(ctx.req.param("status"));
	const delay = Number(ctx.req.query("delay"));
	if (Number.isNaN(status)) {
		return ctx.json({ erro: "status must be a number" });
	}

	console.log("get", status);
	console.log("delay", delay);
	console.log("header", ctx.req.header("teste-header"));
	await sleep(delay);
	return ctx.json({ message: "status" }, status as ContentfulStatusCode, {
		"test-header": ctx.req.header("teste-header") ?? "no-header",
	});
});

app.post("/post/:status", async (ctx) => {
	const status = Number(ctx.req.param("status"));
	if (Number.isNaN(status)) {
		return ctx.json({ erro: "status must be a number" });
	}

	const body = await ctx.req.json();
	console.log("post", status, body);
	return ctx.json(body, status as ContentfulStatusCode);
});

app.put("/put/:status", async (ctx) => {
	const status = Number(ctx.req.param("status"));
	if (Number.isNaN(status)) {
		return ctx.json({ erro: "status must be a number" });
	}

	const body = await ctx.req.json();
	console.log("put", status, body);
	return ctx.json(body, status as ContentfulStatusCode);
});

app.patch("/patch/:status", async (ctx) => {
	const status = Number(ctx.req.param("status"));
	if (Number.isNaN(status)) {
		return ctx.json({ erro: "status must be a number" });
	}

	const body = await ctx.req.json();
	console.log("patch", status, body);
	return ctx.json(body, status as ContentfulStatusCode);
});

export default app;
