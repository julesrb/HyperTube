export type tOauthService = "42" | "github" | "gitlab" | null;

export interface iToken {
    access_token: string
    token_type: "Bearer"
    expires_in: number
}

export interface iUserToken {
    user: iUser
    access_token: string
    refresh_token: string
    token_type: "Bearer"
    expires_in: number
}

export interface iUser {
    id: number
    username: string
    first_name: string
    last_name: string
    email: string
    oauth_method: tOauthService
    color: string
    profile_picture: null | string
    created_at: number
}
