export type tcomment = {
    id: number
    username: string
    comment: string
    created_at: number
};

export const comments: tcomment[] = [{
    id: 1,
    username: "CineMax92",
    comment: "Super film, les effets spéciaux étaient incroyables et l'histoire très prenante.",
    created_at: 1760342400
}, {
    id: 2,
    username: "MovieLover",
    comment: "Un peu long au début mais la fin vaut vraiment le coup.",
    created_at: 1760346000
}, {
    id: 3,
    username: "PopcornAddict",
    comment: "J'ai adoré la bande originale, elle reste en tête toute la journée.",
    created_at: 1760349600
}, {
    id: 4,
    username: "FilmGeek",
    comment: "Très bonne réalisation, mais certains personnages manquaient de profondeur.",
    created_at: 1760353200
}, {
    id: 5,
    username: "NightViewer",
    comment: "Parfait pour une soirée cinéma, je le recommande sans hésiter.",
    created_at: 1760356800
}, {
    id: 6,
    username: "CritiqueDuDimanche",
    comment: "Le scénario est original mais l'exécution aurait pu être meilleure.",
    created_at: 1760360400
}, {
    id: 7,
    username: "SerieFan",
    comment: "Les acteurs jouent très bien, surtout le personnage principal.",
    created_at: 1760364000
}, {
    id: 8,
    username: "CinemaPassion",
    comment: "Un classique instantané, je le reverrai certainement.",
    created_at: 1760367600
}]
