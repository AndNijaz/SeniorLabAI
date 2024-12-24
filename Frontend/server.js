import express from 'express';
import cors from 'cors';
import 'dotenv/config';

const app = express();
const PORT = process.env.REACT_APP_PORT || 4000;

// CORS configuration to allow only https://ai.seniorlab.ba
const corsOptions = {
  origin: '*', // Only allow requests from this origin
    methods: 'POST',                    // Allow only POST method
    allowedHeaders: ['Content-Type'],    // Allow only 'Content-Type' header
    optionsSuccessStatus: 200            // Fallback for older browsers
};

// Apply the CORS middleware with the options
app.use(cors(corsOptions));

// Middleware to parse JSON bodies
app.use(express.json()); 

// Endpoint to handle POST requests from the frontend
app.post('/api/data', async (req, res) => {
    try {
        const userIp = req.headers['x-forwarded-for']?.split(',').shift() || req.socket.remoteAddress;

        const response = await fetch('http://backend:8468/', {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'x-Forwarded-For': userIp
            },
            body: JSON.stringify(req.body)  // Forward the request body
        });

        if (!response.ok) {
            return res.status(response.status).json({ message: 'Error fetching data from backend' });
        }

        const data = await response.json();
        res.json(data);
    } catch (error) {
        console.error('Error:', error);
        res.status(500).json({ message: 'Internal server error', error: error.message });
    }
});

app.listen(PORT, async () => {
    console.log(`Server is running on http://localhost:${PORT}`);
});
