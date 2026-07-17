export type tListResponse<T> = {
    data: T;
    meta?: {
        page: number;
        per_page: number;
        total: number;
    };
};

export type tResponse<T> = {
    data: T;
};

export interface iApplication {
    id: number,
    name: string,
    redirect_uri: string
    client_id: string,
    client_secret: string,
    created_at: string,
    updated_at: string
}
