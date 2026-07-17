export interface iSubtitle {
    id: string
    attributes: {
        subtitle_id: string
        language: string
        files: {
            file_id: number,
        }[]
    }
}

export interface iLoginOpenSubtitles {
    user: {
        allowed_downloads: number
        allowed_translations: number
        level: string,
        user_id: number
        ext_installed: boolean
        vip: boolean
    }
    base_url: string
    token: string
    status: number
}

export interface iDownloadSubtitle {
    link: string
    file_name: string
    requests: number
    remaining: number
    message: string
    reset_time: string
    reset_time_utc: string
}
