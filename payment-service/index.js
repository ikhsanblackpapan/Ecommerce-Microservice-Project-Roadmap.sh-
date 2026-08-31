require("dotenv").config();
const express = require("express");
const jwt = require("jsonwebtoken");
const axios = require("axios");
const { PrismaClient } = require("@prisma/client");
const { PrismaMariaDb } = require("@prisma/adapter-mariadb");

const app = express();
const PORT = process.env.PORT || 8005;
const JWT_SECRET = process.env.JWT_SECRET || "rahasia_super_aman_kamu";

app.use(express.json());

const adapter = new PrismaMariaDb({
  host: process.env.DB_HOST || "db-apps",
  port: 3306,
  user: process.env.DB_USER || "root",
  password: process.env.DB_PASSWORD || "rootpassword",
  database: process.env.DB_NAME || "db_ecommerce_payments",
  connectionLimit: 10,
  allowPublicKeyRetrieval: true,
});

const prisma = new PrismaClient({ adapter });

// Middleware JWT Auth
const authenticateToken = (req, res, next) => {
  const authHeader = req.headers["authorization"];
  const token = authHeader && authHeader.split(" ")[1];

  if (!token) return res.status(401).json({ message: "Token dibutuhkan" });

  jwt.verify(token, JWT_SECRET, (err, user) => {
    if (err) return res.status(403).json({ message: "Token tidak valid" });
    req.user = user;
    req.token = token;
    next();
  });
};

// ==========================================
// ENDPOINT: POST /api/payments/charge
// ==========================================
app.post("/api/payments/charge", authenticateToken, async (req, res) => {
  try {
    const { orderId, paymentMethod, amount } = req.body;

    if (!orderId || !paymentMethod || !amount) {
      return res.status(400).json({ message: "orderId, paymentMethod, dan amount wajib diisi" });
    }

    // 1. Simpan Transaksi Pembayaran ke DB Payment
    const payment = await prisma.payment.create({
      data: {
        orderId,
        userId: req.user.userId,
        amount,
        paymentMethod,
        status: "SUCCESS",
      },
    });

    // 2. Update Status Order menjadi "PAID" di Order Service (Internal Communication)
    const updateOrderUrl = `${process.env.ORDER_SERVICE_URL || "http://order-service:8004"}/api/orders/${orderId}/status`;
    
    await axios.patch(
      updateOrderUrl,
      { status: "PAID" },
      { headers: { Authorization: `Bearer ${req.token}` } }
    );

    return res.status(201).json({
      message: "Pembayaran berhasil diproses!",
      data: payment,
    });
  } catch (error) {
    console.error("Payment Error:", error.response?.data || error.message);
    return res.status(500).json({
      message: "Gagal memproses pembayaran",
      error: error.message,
    });
  }
});

app.listen(PORT, () => {
  console.log(`Payment Service running on port ${PORT}`);
});