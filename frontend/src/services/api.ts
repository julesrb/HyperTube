const API_URL = "http://localhost:8080/api/v1";

export type tListResponse<T> = {
    data: T;
    meta: {
        page: number;
        per_page: number;
        total: number;
    };
};

export type tResponse<T> = {
    data: T;
};

export async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = localStorage.getItem("token");
    const response = await fetch(
        `${API_URL}${endpoint}`,
        {
            ...options,
            headers: {
                "Content-Type": "application/json",
                ...(token && {
                    Authorization: `Bearer ${token}`,
                }),
                ...options.headers,
            },
        }
    );

    if (!response.ok)
        throw new Error("Erreur API");
    return response.json();
}
