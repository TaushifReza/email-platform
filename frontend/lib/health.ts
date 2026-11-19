// Health check function that calls the backend
export async function checkHealth(): Promise<{ status: string }> {
  try {
    const response = await fetch('/api/healthz');
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    console.error('Health check failed:', error);
    throw error;
  }
}