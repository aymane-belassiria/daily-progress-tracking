export function buildGraphLayout(nodes = []) {
  const width = 720;
  const rowHeight = 104;
  const columns = Math.min(4, Math.max(1, nodes.length));
  const columnWidth = width / columns;

  const positioned = nodes.map((node, index) => {
    const column = index % columns;
    const row = Math.floor(index / columns);
    return {
      ...node,
      x: Math.round(columnWidth * column + columnWidth / 2),
      y: 58 + row * rowHeight
    };
  });

  const edges = positioned.flatMap((node) =>
    (node.depends_on || []).map((dependency) => ({
      from: dependency,
      to: node.day_index
    }))
  );

  return {
    width,
    height: Math.max(180, 116 + Math.ceil(positioned.length / columns) * rowHeight),
    nodes: positioned,
    edges
  };
}

export function scoreLabel(score = 0) {
  if (score >= 80) {
    return "Strong";
  }
  if (score >= 50) {
    return "Building";
  }
  return "Starting";
}

export function buildScoreTrend(entries = [], roadmap) {
  const value = roadmap?.score?.overall || 0;
  const recent = entries
    .slice(0, 7)
    .sort((left, right) => left.entry_date.localeCompare(right.entry_date));
  if (recent.length === 0) {
    return [{ label: "Today", value }];
  }
  return recent.map((entry) => ({
    label: entry.entry_date.slice(5),
    value
  }));
}

export function taskTotals(roadmap) {
  const tasks = (roadmap?.nodes || []).flatMap((node) => node.tasks || []);
  const done = tasks.filter((task) => task.done).length;
  return { done, total: tasks.length };
}
