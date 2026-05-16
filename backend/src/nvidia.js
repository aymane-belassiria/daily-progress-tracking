const NVIDIA_URL = "https://integrate.api.nvidia.com/v1/chat/completions";

function extractText(data) {
  const choice = data?.choices?.[0];
  return (
    choice?.message?.content ||
    choice?.delta?.content ||
    "No response returned by the language model."
  );
}

export async function generateWithNvidia({ apiKey, model, system, prompt }) {
  const response = await fetch(NVIDIA_URL, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${apiKey}`
    },
    body: JSON.stringify({
      model,
      messages: [
        { role: "system", content: system },
        { role: "user", content: prompt }
      ],
      temperature: 0.6,
      top_p: 0.7,
      max_tokens: 4096,
      stream: false
    })
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(`NVIDIA API error ${response.status}: ${message}`);
  }

  const data = await response.json();
  return extractText(data);
}
