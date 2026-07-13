// index.js
// Example Express application importing and using the Limiter.io Node.js SDK

const express = require('express');
const LimiterClient = require('../client');
const expressRateLimit = require('../middleware');

const app = express();
app.use(express.json());

// 1. Initialize the official SDK client
const client = new LimiterClient('http://localhost:8080', 'replace_with_developer_api_key');

// 2. Mount endpoints and apply the Express SDK middleware
app.get('/api/payments', expressRateLimit(client), (req, res) => {
  res.json({
    status: 'success',
    message: 'Payment successfully processed using Node.js SDK!'
  });
});

app.get('/api/data', expressRateLimit(client), (req, res) => {
  res.json({
    status: 'success',
    message: 'Telemetry metrics loaded using Node.js SDK.'
  });
});

const PORT = process.env.PORT || 9001;
app.listen(PORT, () => {
  console.log(`🚀 Node.js Client App running on http://localhost:${PORT}`);
});
