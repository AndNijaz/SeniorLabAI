export const fetchQuery = async (text: string) => {
  try {
    // Use the correct URL to point at your Express server
    const endpoint = import.meta.env.REACT_APP_API_ENDPOINT || 'http://132.226.195.28:4000/api/data';

    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text })
    });

    if (!response.ok) {
      throw new Error('Failed to fetch data');
    }

    return await response.json();
  } catch (error) {
    console.error('Error:', error);
    throw error;
  }
};