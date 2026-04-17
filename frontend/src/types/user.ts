export type tuser = {
    id: number
    username: string
    firstname: string
    lastname: string
    email: string
    color: string
    profile_picture: null | string
    film_history: number[]
    joined_at: number
};

export const users: tuser[] = [
    {
        id: 0,
        username: "fguirama",
        firstname: "Florian",
        lastname: "Guiramand",
        email: "florian.guiramand@example.com",
        color: "purple",
        profile_picture: "/images/profile_pictures.jpeg",
        film_history: [0, 1, 2, 3, 4, 5, 6, 7],
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
        film_history: [],
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
        film_history: [],
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
        film_history: [],
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
        film_history: [],
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
        film_history: [],
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
        film_history: [],
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
        film_history: [],
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
        film_history: [],
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
        film_history: [],
        joined_at: 1749542400
    }
];
