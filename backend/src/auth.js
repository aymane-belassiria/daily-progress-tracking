import jwt from "jsonwebtoken";

export function issueToken(email, secret) {
  return jwt.sign({ sub: email }, secret, { expiresIn: "7d" });
}

export function verifyToken(token, secret) {
  return jwt.verify(token, secret);
}

export function requireAuth(secret) {
  return (req, res, next) => {
    const header = req.headers.authorization || "";
    const token = header.startsWith("Bearer ") ? header.slice(7) : null;

    if (!token) {
      return res.status(401).json({ error: "Missing bearer token." });
    }

    try {
      const payload = verifyToken(token, secret);
      req.user = { email: payload.sub };
      next();
    } catch {
      res.status(401).json({ error: "Invalid or expired token." });
    }
  };
}
