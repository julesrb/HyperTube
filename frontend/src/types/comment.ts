export type tComment = {
    id: number
    author_id: number
    author_username: string
    author_firstname: string
    author_lastname: string
    author_profile_pictures: null | string
    author_color: string
    comment: string
    edited: boolean
    created_at: number
};

export const comments: tComment[] = [{
    id: 0,
    author_id: 0,
    author_username: "fguirama",
    author_firstname: "fguirama",
    author_lastname: "fguirama",
    author_profile_pictures: "/images/profile_pictures.jpeg",
    author_color: "yellow",
    comment: "Super film, les effets spéciaux étaient incroyables et l'histoire très prenante.",
    edited: false,
    created_at: 1765142400
}, {
    id: 1,
    author_id: 2,
    author_username: "codewizard",
    author_firstname: "codewizard",
    author_lastname: "codewizard",
    author_profile_pictures: null,
    author_color: "yellow",
    comment: "Un peu long au début mais la fin vaut vraiment le coup.",
    edited: false,
    created_at: 1720346000
}, {
    id: 2,
    author_id: 2,
    author_username: "nightcoder",
    author_firstname: "nightcoder",
    author_lastname: "nightcoder",
    author_profile_pictures: null,
    author_color: "yellow",
    comment: "J'ai adoré la bande originale, elle reste en tête toute la journée.",
    edited: false,
    created_at: 1750349600
}, {
    id: 3,
    author_id: 2,
    author_username: "designqueen",
    author_firstname: "designqueen",
    author_lastname: "designqueen",
    author_profile_pictures: null,
    author_color: "yellow",
    comment: "Très bonne réalisation, mais certains personnages manquaient de profondeur.",
    edited: true,
    created_at: 1760353200
}, {
    id: 4,
    author_id: 2,
    author_username: "bugslayer",
    author_firstname: "bugslayer",
    author_lastname: "bugslayer",
    author_profile_pictures: null,
    author_color: "yellow",
    comment: "Parfait pour une soirée cinéma, je le recommande sans hésiter.",
    edited: false,
    created_at: 1768356800
}, {
    id: 5,
    author_id: 2,
    author_username: "fastdeploy",
    author_firstname: "fastdeploy",
    author_lastname: "fastdeploy",
    author_profile_pictures: null,
    author_color: "yellow",
    comment: "Le scénario est original mais l'exécution aurait pu être meilleure.",
    edited: true,
    created_at: 1760360400
}, {
    id: 6,
    author_id: 2,
    author_username: "terminalguru",
    author_firstname: "terminalguru",
    author_lastname: "terminalguru",
    author_profile_pictures: null,
    author_color: "yellow",
    comment: "Les acteurs jouent très bien, surtout le personnage principal.",
    edited: false,
    created_at: 1730364000
}, {
    id: 7,
    author_id: 0,
    author_username: "fguirama",
    author_firstname: "fguirama",
    author_lastname: "fguirama",
    author_profile_pictures: null,
    author_color: "yellow",
    comment: "Sublime ! Deux plans résument à merveille ce film :" +
        "Le premier plan qui m'a marqué, c'est celui où la mère est en entretien pour renouveler sa période d’essai. Le plan film deux personnages, l'une en face de l'autre. À droite, la femme de ménage, âgée, abîmée par la vie, par son travail, vêtue de sa blouse de travail bleu clair. Dans le fond, on y voit des casiers, des formes géométriques délimitées qui emprisonnent le personnage. Elle rentre dans les cases. À gauche, en opposition, on y retrouve le personnage interprété par Camélia Jordana, elle est jeune et bien habillée. Dans le fond, une fenêtre, montrant sa liberté. Elle ne veut pas rentrer dans une case : ni celle de femme, ni celle \"d'arabe\". Elle se veut libre. Et ça, ça pose un problème à la société, surtout celle de l'époque. Le film montre assez justement la volonté de cette mère, ne souhaitant correspondre à aucun stéréotype. Le film va même admettre les quelques excès du personnage, permettant de mieux nuancer son propos.\n" +
        "Le second plan très intéressant, c'est celui de l'autre personnage principal, Mouna, sa fille. On la voit dans une salle de classe vide (excepté ses parents et la professeure, pas présents dans le cadre). Elle est devant le tableau, mais pas tout à fait au centre, légèrement désaxée sur la droite. Des ombres de volets à lamelles se projettent sur son visage. Elle est emprisonnée. La dualité entre son histoire personnelle et celle de son pays, la France, la questionne, la tourmente. Puis, on voit Mouna s'avancer légèrement, sortir de l'ombre, et venir se mettre au centre du tableau pour réciter un poème. Il marque un tournant dans le scénario et dans le développement du personnage. Cette scène va ensuite faire écho avec une autre plus tard, qu'on ne verra pas. On entend simplement Mouna réciter son exposé sur Charles Martel, la réécriture historique de certains événements et la récupération politique qui en a suivi.\n" +
        "Le film est beau et touchant. Camélia Jordan et Sofiane Zermani (fianso) sont merveilleux. L'actrice de Mouna l'est aussi.",
    edited: true,
    created_at: 1776105275
}]
