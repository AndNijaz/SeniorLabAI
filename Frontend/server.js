import express from 'express';
import cors from 'cors';
import 'dotenv/config';

const app = express();
const PORT = process.env.REACT_APP_PORT || 4000;

// CORS configuration to allow only http://132.226.195.28
const corsOptions = {
    origin: 'http://132.226.195.28:9911',  // Allow only this specific origin
    optionsSuccessStatus: 200 // For legacy browsers
  };

  app.use(cors(corsOptions)); // Apply the CORS middleware with the options
  app.use(express.json());
  
  // Endpoint to handle requests from the frontend
  app.post('/api/data', async (req, res) => {
    try {
      const userIp = req.headers['x-forwarded-for'] || req.socket.remoteAddress;
      const response = await fetch('http://backend:8468/', {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'x-Forwarded-For': userIp
         },
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

  app.listen(PORT, async () => {
    console.log(`Server is running on http://localhost:${PORT}`);
    
  });
