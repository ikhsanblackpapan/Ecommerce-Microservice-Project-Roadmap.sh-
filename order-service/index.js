require("dotenv").config();
const express = require("express");
const jwt = require("jsonwebtoken");
const axios = require("axios");
const { PrismaClient } = require("@prisma/client");
const { PrismaMariaDb } = require("@prisma/adapter-mariadb");
const amqplib = require("amqplib");

const app = express();
const PORT = process.env.PORT || 8004;
const JWT_SECRET = process.env.JWT_SECRET || "rahasia_super_aman_kamu";

app.use(express.json());

const adapter = new PrismaMariaDb({
  host: process.env.DB_HOST || "db-apps",
  port: 3306,
  user: process.env.DB_USER || "root",
  password: process.env.DB_PASSWORD || "rootpassword",
  database: process.env.DB_NAME || "db_ecommerce_orders",
  connectionLimit: 10,
  allowPublicKeyRetrieval: true,
});


const prisma = new PrismaClient({ adapter });

//jwt midleware

const authenticateToken = (req, res, next) => {
  const authHeader = req.headers["authorization"];
  const token = authHeader && authHeader.split(" ")[1];

  if (!token) {
    return res.status(401).json({ message: "Authorization token required" });
  }

  jwt.verify(token, JWT_SECRET, (err, user) => {
    if (err) {
      return res.status(403).json({ message: "Invalid or expired token" });
    }
    req.user = user;
    req.token = token;
    next();
  });
};

let channel, connection;

// async function connectRabbitMQ() {
async function connectRabbitMQ() {
  try {
    // conect to RabbitMQ server
    const amqpServer = process.env.RABBITMQ_URL || "amqp://localhost:5672";
    connection = await amqplib.connect(amqpServer);
    channel = await connection.createChannel();


    await channel.assertQueue("ORDER_CREATED_QUEUE", { durable: true });
    console.log("Connected to RabbitMQ");
  } catch (error) {
    console.error("Failed to connect to RabbitMQ:", error);
    setTimeout(connectRabbitMQ, 5000); 
  }
   
  }
  connectRabbitMQ();




// post /api/orders/checkout
app.post("/api/orders/checkout", authenticateToken, async (req, res) => {
  try {
    const userId = req.user.userId;

    const cartResponse = await axios.get( 
      `${process.env.CART_SERVICE_URL || "http://cart-service:8003"}/api/cart`,
      {
        headers: { Authorization: `Bearer ${req.token}` },
      }
    );

    const cartItems = cartResponse.data.items;

    if (!cartItems || cartItems.length === 0) {
      return res.status(400).json({ message: "keranjang belanja kosong " });
    }

    const totalPrice = cartItems.reduce((acc, item) => acc + item.price * item.quantity, 0);

      const newOrder = await prisma.order.create({
        data: {
          userId: userId,
          totalPrice,
          status: "PENDING",
          items: {
            create: cartItems.map((item) => ({
              productId: item.product_id,
              productName: item.name,
              price: item.price,
              quantity: item.quantity,
            })),
          },
        },

        include: {
          items: true,
        },
    });

    const eventPayLoad = {
      event: "ORDER_CREATED",
      user_id: userId,
      orderId: newOrder.id,
      token: req.token,
    };

    if (channel) {
      channel.sendToQueue("ORDER_CREATED_QUEUE", Buffer.from(JSON.stringify(eventPayLoad)));
      console.log("Published ORDER_CREATED event sent to RabbitMQ");
  
    }


    const clearCartUrl = `${process.env.CART_SERVICE_URL || "http://cart-service:8003"}/api/cart/clear`;
    await axios.delete(clearCartUrl, {
      headers: { Authorization: `Bearer ${req.token}` },
    });

    return res.status(201).json({
      message: "Order berhasil dibuat",
      data: newOrder,
    });
  } catch (error) {
    console.error("Error checkout:", error.response?.data || error.message);
    return res.status(500).json({
      message: "Gagal memproses checkout",
      error: error.message,
    });
  }
});

// endpoint GET /api/orders/ ( LIHAT RIWAYAT ORDER )

app.get("/api/orders", authenticateToken, async (req, res) => {
  try {
    const userId = req.user.userId;

    const orders = await prisma.order.findMany({
      where: { userId },
      include: { items: true },
      orderBy: { createdAt: "desc" },
    });

    return res.json({
      message: "Riwayat order berhasil diambil",
      data: orders,
    });
  } catch (error) {
    console.error(error);
    return res.status(500).json({
      message: "Internal server error",
    });
  }
});

//mengubah status pembayaran 
// async
app.patch("/api/orders/:id/status", authenticateToken, async (req, res) => {
  try {
    const { id } = req.params;
    const { status } = req.body;

    const updatedOrder = await prisma.order.update({
      where: { id },
      data: { status },
    });

    return res.json({
      message: "Status order berhasil diperbarui",
      data: updatedOrder,
    });
  } catch (error) {
    return res.status(500).json({ message: "Gagal update status order", error: error.message });
  }
});


app.listen(PORT, () => {
  console.log(`Order service is running on port ${PORT}`);
});
