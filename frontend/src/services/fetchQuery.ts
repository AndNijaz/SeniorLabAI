export const fetchQuery = async (text: string) => {
  console.log("aaaaa " + text);
  // eslint-disable-next-line no-useless-catch
  try {
    // Instead of URLSearchParams, construct a JSON object
    const response = await fetch(
      import.meta.env.VITE_OPEN_AI_ENDPOINT as string,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json", // Set the Content-Type to application/json
        },
        body: JSON.stringify({ text: text }), // Send text in JSON format
      }
    );

    if (!response.ok) {
      throw new Error("Failed to fetch data");
    }

    const data = await response.json();
    return data;
  } catch (error) {
    throw error;
  }
};
