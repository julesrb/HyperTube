export interface iUser {
    id: number
    username: string
    firstname: string
    lastname: string
    email: string
    color: string
    profile_picture: null | string
    watch_history: {movie_id: string, watch_percent: number}[]
    joined_at: number
}

export const users: iUser[] = [
    {
        id: 0,
        username: "fguirama",
        firstname: "Florian",
        lastname: "Guiramand",
        email: "florian.guiramand@example.com",
        color: "purple",
        profile_picture: "/images/profile_pictures.jpeg",
        watch_history: [
            { movie_id: "tt36856791", watch_percent: 34 },
            { movie_id: "tt39749979", watch_percent: 100 },
            { movie_id: "tt32560777", watch_percent: 25 },
            { movie_id: "tt28650488", watch_percent: 55 },
            { movie_id: "tt6105098", watch_percent: 100 },
            { movie_id: "tt34794183", watch_percent: 62 },
        ],
        joined_at: 1746764800
    },
    {
        id: 1,
        username: "codewizard",
        firstname: "Emma",
        lastname: "Bernard",
        email: "emma.bernard@example.com",
        color: "pink",
        profile_picture: null,
        watch_history: [],
        joined_at: 1748851200
    },
    {
        id: 2,
        username: "nightcoder",
        firstname: "Hugo",
        lastname: "Dubois",
        email: "hugo.dubois@example.com",
        color: "pink",
        profile_picture: null,
        watch_history: [],
        joined_at: 1748937600
    },
    {
        id: 3,
        username: "designqueen",
        firstname: "Chloé",
        lastname: "Moreau",
        email: "chloe.moreau@example.com",
        color: "blue",
        profile_picture: null,
        watch_history: [],
        joined_at: 1749024000
    },
    {
        id: 4,
        username: "bugslayer",
        firstname: "Nathan",
        lastname: "Laurent",
        email: "nathan.laurent@example.com",
        color: "yellow",
        profile_picture: null,
        watch_history: [],
        joined_at: 1749110400
    },
    {
        id: 5,
        username: "fastdeploy",
        firstname: "Sarah",
        lastname: "Simon",
        email: "sarah.simon@example.com",
        color: "purple",
        profile_picture: null,
        watch_history: [],
        joined_at: 1749196800
    },
    {
        id: 6,
        username: "terminalguru",
        firstname: "Thomas",
        lastname: "Michel",
        email: "thomas.michel@example.com",
        color: "blue",
        profile_picture: null,
        watch_history: [],
        joined_at: 1749283200
    },
    {
        id: 7,
        username: "datastream",
        firstname: "Julie",
        lastname: "Garcia",
        email: "julie.garcia@example.com",
        color: "red",
        profile_picture: null,
        watch_history: [],
        joined_at: 1749369600
    },
    {
        id: 8,
        username: "cloudrider",
        firstname: "Alexandre",
        lastname: "Roux",
        email: "alexandre.roux@example.com",
        color: "green",
        profile_picture: null,
        watch_history: [],
        joined_at: 1749456000
    },
    {
        id: 9,
        username: "scriptmaster",
        firstname: "Manon",
        lastname: "Fournier",
        email: "manon.fournier@example.com",
        color: "purple",
        profile_picture: null,
        watch_history: [],
        joined_at: 1749542400
    }
];
