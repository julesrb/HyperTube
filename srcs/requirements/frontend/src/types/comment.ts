import {iUser} from "@/types/user";
import {iMovie} from "@/types/movie";

export interface iComment {
    id: number
    user: iUser
    content: string
    edited: boolean
    updated_at: number
}

export interface iCommentDetails extends iComment {
    movie: iMovie
}
