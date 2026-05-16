import test from "node:test";
import assert from "node:assert/strict";

import { resolveApiBaseUrl, toApiUrl } from "./api.js";
import { buildGraphLayout, buildScoreTrend, scoreLabel } from "./roadmap.js";

test("resolveApiBaseUrl trims a trailing slash", () => {
  assert.equal(resolveApiBaseUrl("https://api.example.com/"), "https://api.example.com");
});

test("resolveApiBaseUrl falls back to localhost", () => {
  assert.equal(resolveApiBaseUrl(""), "http://localhost:4000");
});

test("toApiUrl joins the normalized base URL with API paths", () => {
  assert.equal(toApiUrl("https://api.example.com/", "/api/health"), "https://api.example.com/api/health");
});

test("buildGraphLayout positions roadmap nodes in stable rows", () => {
  const layout = buildGraphLayout([
    { id: 1, day_index: 1, title: "One", depends_on: [] },
    { id: 2, day_index: 2, title: "Two", depends_on: [1] }
  ]);

  assert.equal(layout.width, 720);
  assert.equal(layout.nodes.length, 2);
  assert.deepEqual(layout.edges, [{ from: 1, to: 2 }]);
  assert.ok(layout.nodes[1].x > layout.nodes[0].x);
});

test("scoreLabel explains score bands", () => {
  assert.equal(scoreLabel(84), "Strong");
  assert.equal(scoreLabel(62), "Building");
  assert.equal(scoreLabel(25), "Starting");
});

test("buildScoreTrend extracts dated score points", () => {
  const trend = buildScoreTrend([
    { entry_date: "2026-05-16" },
    { entry_date: "2026-05-17" }
  ], { score: { overall: 70 } });

  assert.equal(trend.length, 2);
  assert.deepEqual(trend[1], { label: "05-17", value: 70 });
});
