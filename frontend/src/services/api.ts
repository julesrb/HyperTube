const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function apiFetch(
    endpoint: string,
    options: RequestInit = {}
) {
    const response = await fetch(
        `${API_URL}${endpoint}`,
        {
            ...options,
            headers: {
                "Content-Type": "application/json",
                ...options.headers,
            },
            credentials: "include",
        }
    );

    if (!response.ok) {
        throw new Error("API error");
    }

    return response.json();
}
