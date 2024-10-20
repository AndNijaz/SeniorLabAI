import express from 'express';
import 'dotenv/config';

const app = express();
const PORT = process.env.REACT_APP_PORT || 4000;

app.use(express.json());

// Endpoint to handle requests from the frontend
app.post('/api/data', async (req, res) => {
  try {
    const response = await fetch(process.env.VITE_OPEN_AI_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req.body)  // Forward the request body
    });

    if (!response.ok) {
      return res.status(response.status).json({ message: 'Error fetching data' });
    }

    const data = await response.json();
    res.json(data);
  } catch (error) {
    res.status(500).json({ message: 'Internal server error', error: error.message });
  }
});

app.listen(PORT, () => {
  console.log(`Server is running on http://localhost:${PORT}`);
});
