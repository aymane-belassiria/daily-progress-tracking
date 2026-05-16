import "dotenv/config";
import bcrypt from "bcryptjs";
import cors from "cors";
import express from "express";
import { z } from "zod";
import { requireAuth, issueToken } from "./auth.js";
import {
  createGoal,
  deleteGoal,
  getDashboard,
  listEntries,
  listGoals,
  upsertEntry,
  updateGoal
} from "./db.js";
import { generateWithNvidia } from "./nvidia.js";
import { hintsPrompt, roadmapPrompt } from "./prompts.js";

const app = express();
const port = Number(process.env.PORT || 4000);
const jwtSecret = process.env.JWT_SECRET;
const adminEmail = process.env.ADMIN_EMAIL;
const adminPassword = process.env.ADMIN_PASSWORD;
const frontendOrigin = process.env.FRONTEND_ORIGIN || "*";
const nvidiaApiKey = process.env.NVIDIA_API_KEY;
const nvidiaModel = process.env.NVIDIA_MODEL || "qwen/qwen3-next-80b-a3b-instruct";

if (!jwtSecret || !adminEmail || !adminPassword) {
  throw new Error("Missing required auth configuration in backend/.env.");
}

const adminPasswordHash = bcrypt.hashSync(adminPassword, 10);

app.use(
  cors({
    origin: frontendOrigin === "*" ? true : frontendOrigin,
    credentials: false
  })
);
app.use(express.json({ limit: "1mb" }));

const authMiddleware = requireAuth(jwtSecret);

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8)
});

const goalSchema = z.object({
  period: z.enum(["weekly", "monthly"]),
  title: z.string().min(1).max(160),
  description: z.string().max(2000).default(""),
  target_value: z.coerce.number().int().min(1).max(100000).default(1),
  current_value: z.coerce.number().int().min(0).max(100000).default(0),
  due_date: z.string().max(20).optional().nullable(),
  status: z.enum(["active", "completed", "paused"]).default("active")
});

const entrySchema = z.object({
  entry_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
  summary: z.string().max(2000).default(""),
  wins: z.string().max(2000).default(""),
  blockers: z.string().max(2000).default(""),
  notes: z.string().max(4000).default(""),
  mood: z.string().max(80).default("")
});

const aiSchema = z.object({
  prompt: z.string().max(2000).optional().default("")
});

app.get("/api/health", (_req, res) => {
  res.json({ ok: true });
});

app.post("/api/auth/login", (req, res) => {
  const parsed = loginSchema.safeParse(req.body);

  if (!parsed.success) {
    return res.status(400).json({ error: "Invalid login payload." });
  }

  const { email, password } = parsed.data;
  const emailMatches = email.toLowerCase() === adminEmail.toLowerCase();
  const passwordMatches = bcrypt.compareSync(password, adminPasswordHash);

  if (!emailMatches || !passwordMatches) {
    return res.status(401).json({ error: "Invalid credentials." });
  }

  res.json({
    token: issueToken(adminEmail, jwtSecret),
    user: { email: adminEmail }
  });
});

app.get("/api/auth/me", authMiddleware, (req, res) => {
  res.json({ user: req.user });
});

app.get("/api/dashboard", authMiddleware, (_req, res) => {
  res.json(getDashboard());
});

app.get("/api/goals", authMiddleware, (req, res) => {
  const period = req.query.period;

  if (period && period !== "weekly" && period !== "monthly") {
    return res.status(400).json({ error: "Invalid period." });
  }

  res.json({ goals: listGoals(period) });
});

app.post("/api/goals", authMiddleware, (req, res) => {
  const parsed = goalSchema.safeParse(req.body);

  if (!parsed.success) {
    return res.status(400).json({ error: "Invalid goal payload.", details: parsed.error.flatten() });
  }

  res.status(201).json({ goal: createGoal(parsed.data) });
});

app.put("/api/goals/:id", authMiddleware, (req, res) => {
  const parsed = goalSchema.safeParse(req.body);

  if (!parsed.success) {
    return res.status(400).json({ error: "Invalid goal payload.", details: parsed.error.flatten() });
  }

  const goal = updateGoal(Number(req.params.id), parsed.data);
  if (!goal) {
    return res.status(404).json({ error: "Goal not found." });
  }

  res.json({ goal });
});

app.delete("/api/goals/:id", authMiddleware, (req, res) => {
  deleteGoal(Number(req.params.id));
  res.status(204).end();
});

app.get("/api/entries", authMiddleware, (req, res) => {
  const limit = Number(req.query.limit || 30);
  res.json({ entries: listEntries(Math.max(1, Math.min(limit, 120))) });
});

app.post("/api/entries", authMiddleware, (req, res) => {
  const parsed = entrySchema.safeParse(req.body);

  if (!parsed.success) {
    return res.status(400).json({ error: "Invalid entry payload.", details: parsed.error.flatten() });
  }

  res.json({ entry: upsertEntry(parsed.data) });
});

app.post("/api/ai/roadmap", authMiddleware, async (req, res) => {
  if (!nvidiaApiKey) {
    return res.status(500).json({ error: "NVIDIA_API_KEY is not configured." });
  }

  const parsed = aiSchema.safeParse(req.body);
  if (!parsed.success) {
    return res.status(400).json({ error: "Invalid AI request payload." });
  }

  try {
    const dashboard = getDashboard();
    const content = await generateWithNvidia({
      apiKey: nvidiaApiKey,
      model: nvidiaModel,
      system: "You create realistic execution roadmaps for one private user.",
      prompt: roadmapPrompt(dashboard.goals, dashboard.entries, parsed.data.prompt)
    });

    res.json({ content });
  } catch (error) {
    res.status(502).json({ error: error.message });
  }
});

app.post("/api/ai/hints", authMiddleware, async (req, res) => {
  if (!nvidiaApiKey) {
    return res.status(500).json({ error: "NVIDIA_API_KEY is not configured." });
  }

  const parsed = aiSchema.safeParse(req.body);
  if (!parsed.success) {
    return res.status(400).json({ error: "Invalid AI request payload." });
  }

  try {
    const dashboard = getDashboard();
    const content = await generateWithNvidia({
      apiKey: nvidiaApiKey,
      model: nvidiaModel,
      system: "You create short, useful guidance for one private user.",
      prompt: hintsPrompt(dashboard.goals, dashboard.entries, parsed.data.prompt)
    });

    res.json({ content });
  } catch (error) {
    res.status(502).json({ error: error.message });
  }
});

app.listen(port, () => {
  console.log(`Backend listening on http://localhost:${port}`);
});
