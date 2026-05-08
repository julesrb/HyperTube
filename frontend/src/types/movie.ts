export interface iMovie {
    imdb_id: string
    title: string
    year: string
    poster_url: string
    backdrop_url: string
    genres: number[]
    rate: number
}

export interface iMovieDetails extends iMovie {
    "tmdb_id": string
    "note": number
    "genres": number[]
    "runtime_minutes": number
    "summary": string
    "director": string
    "cast": string[]
    "watched": boolean
    "progression": number// todo handle change 0.0
}

export interface iGenre {
    "id": number
    "name": string
}
