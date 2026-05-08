import {iUser} from "@/types/user";

export interface iComment {
    id: number
    movie_id: string
    user: iUser
    content: string
    edited: boolean
    updated_at: number
}
