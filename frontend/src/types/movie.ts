export type tmovies = {
    id: number
    title: string
    src: string
    year: string
    backdrops: string[]
    synopsis: string
    genres: string[]
    directors: string[]
    stars: string[]
    length: string
};

export const movies: tmovies[] = [{
    id: 1,
    title: "Avatar: Fire and Ash",
    src: "dVv28yL7kyeMK3sUTWVSKrZC6tV.jpg",
    year: "2025",
    backdrops: ["3Dqievkc7krcTtDE2hjRkIsEzB1.webp", "iN41Ccw4DctL8npfmYg1j5Tr1eb.webp", "sdZSjtGUTSN8B3al5o0f2WoQfQQ.webp", "vm4H1DivjQoNIm0Vs6i3CTzFxQ0.webp"],
    synopsis: "Après la mort de Neteyam, Jake et Neytiri affrontent leur chagrin tout en faisant face au Peuple des Cendres, une tribu Na’vi redoutable menée par la fougueuse Varang, alors que le conflit sur Pandora s’intensifie et qu’une nouvelle quête morale s’amorce.",
    genres: ["Science-Fiction", "Aventure", "Fantastique"],
    directors: ["Sam Worthington", "Zoe Saldaña", "Sigourney Weaver", "Stephen Lang"],
    stars: ["James Cameron"],
    length: "3h 18m"
}, {
    id: 2,
    title: "The Drama",
    src: "heV850U3uf2FNxLW8QPcKyZMfdx.jpg",
    year: "2026",
    backdrops: ["1oKLEA9JOhvaBwLpqjROisvWMy7.webp", "gQxnJoswBzsENExzygVnFiemvK7.webp", "rw1WmxUvmbwRsbAX12m6toIg04j.webp", "t1S7cLojbI92zEkdTgkWnytW7Oa.webp", "vc7TeZyKhBeVA4YI3z1KAeIXqGV.webp", "yLjh1d3SlneSRVHz0ZrX1fuXSob.webp"],
    synopsis: "Dans les jours précédant leur mariage, un couple est confronté à une crise lorsque des révélations inattendues font dérailler ce que l'un d'eux pensait savoir de l'autre.",
    genres: ["Romance", "Comédie", "Drame"],
    directors: ["Kristoffer Borgli"],
    stars: ["Zendaya", "Robert Pattinson"],
    length: "1h 45m"
}, {
    id: 3,
    title: "Project Hail Mary",
    src: "kutdTXxqWJ7KYmnll8nSrtWWsCR.jpg",
    year: "2026",
    backdrops: ["2I1OFQJ0L9T0dpU6FobKFWV2PxX.webp", "8Tfys3mDZVp4tNoH2ktm06a0Tau.webp"],
    synopsis: "Le professeur de sciences Ryland Grace se réveille à bord d’un vaisseau spatial, à des années-lumière de la Terre, sans aucun souvenir de qui il est ni de la façon dont il est arrivé là. À mesure que sa mémoire lui revient, il commence à découvrir sa mission : résoudre l’énigme d’une mystérieuse substance qui provoque l’extinction du Soleil. Il doit faire appel à ses connaissances scientifiques et ses idées peu orthodoxes pour sauver toute vie sur Terre mais une amitié inattendue pourrait signifier qu’il n’aura pas à accomplir cette tâche seul.",
    genres: ["Science-Fiction", "Aventure"],
    directors: ["Phil Lord", "Christopher Miller"],
    stars: ["Ryan Gosling", "James Ortiz", "Sandra Hüller"],
    length: "2h 35m"
}, {
    id: 4,
    title: "The Super Mario Galaxy Movie",
    src: "6RmUnKyBfLu6MfC7sya2WXjyVut.jpg",
    year: "2026",
    backdrops: ["9Z2uDYXqJrlmePznQQJhL6d92Rq.webp", "kxQiIJ4gVcD3K6o14MJ72p5yRcE.webp", "xTd74mchq96pqfdlgmKDMQx12gv.webp"],
    synopsis: "Au-delà du Royaume Champignon, Mario, Luigi et Yoshi voyagent à travers les galaxies aux côtés de Rosalina pour arrêter Bowser Jr., dont la tentative de sauver son père menace l’équilibre de l’univers.",
    genres: ["Familial", "Comédie", "Aventure", "Fantastique", "Animation"],
    directors: ["Michael Jelenic", "Aaron Horvath"],
    stars: ["Chris Pratt", "Charlie Day", "Anya Taylor-Joy"],
    length: "1h 40m"
}, {
    id: 5,
    title: "Hoppers",
    src: "p5G1Fbn6hjQprNijkcjarf8FZWG.jpg",
    year: "2026",
    backdrops: ["2RrLuIfIzGWWIH8IAEo6o0IYHmx.webp", "7Zk07DUBunUp9E1LtLK63jJGiDk.webp", "gTKBGnxWVVg3yHA3mjJjnJRgU1v.webp", "u53UYu5XG2hNgWGvs3xGhAVzypl.webp"],
    synopsis: "Mabel est une jeune fille audacieuse qui a une grande ambition : infiltrer le monde animal. Pour ce faire, elle transfère son esprit à un robot castor afin de pouvoir approcher d’autres animaux et étudier leur mode de vie.",
    genres: ["Animation", "Familial", "Science-Fiction", "Comédie", "Aventure"],
    directors: ["Daniel Chong"],
    stars: ["Piper Curda", "Bobby Moynihan", "Jon Hamm"],
    length: "1h 45m"
}, {
    id: 6,
    title: "The Bride!",
    src: "g7twdxxP2QuWLE8hOgSl5570WoX.jpg",
    year: "2026",
    backdrops: ["7LivZ6bgoG7JQ6c4wJCQmyjj6ye.webp", "l8rKKMU2M9dDULO9CEtDNdWAEUJ.webp", "uXgWbwxmCW6XZ3ylRThk9Sqhwwq.webp"],
    synopsis: "Un Frankenstein esseulé se rend dans le Chicago des années 30 afin de convaincre la brillante scientifique Dr Euphronious de lui concevoir une compagne. Les deux ressuscitent une jeune femme assassinée et La fiancée voit le jour. Et ce qui suit dépasse toutes leurs attentes : Meurtre ! Possession ! Un mouvement culturel déchaîné et radical ! Et des amants rebelles dans une romance sauvage et explosive !",
    genres: ["Science-Fiction", "Horreur", "Fantastique"],
    directors: ["Maggie Gyllenhaal"],
    stars: ["Jessie Buckley", "Christian Bale"],
    length: "2h 7m"
}, {
    id: 7,
    title: "Send Help",
    src: "iXW9WMVOjhJ3wwOGz5Da42tU4wk.jpg",
    year: "2026",
    backdrops: ["8zeiIiI71I93ZnSTkda3WfrLxki.webp", "493MuzSs6iLyKM7aqsA30eODfsV.webp", "gCmfeKmEAZBP5gcXpiqb0gii9rS.webp", "hO2jx1H3XafR7Y8QbFgVH1sHTY9.webp"],
    synopsis: "",
    genres: ["Horreur", "Thriller", "Comédie"],
    directors: ["Sam Raimi"],
    stars: ["Rachel McAdams", "Dylan O'Brien"],
    length: "1h 53m"
}, {
    id: 8,
    title: "GOAT",
    src: "aGS0uR26XkSkkc0Grxu4ZrFEVu5.webp",
    year: "2026",
    backdrops: ["6qMyijXWuDfrN8bXUJT1DErjQog.webp", "imkZLxabgZBdsWPnkPkFCHC9NCw.webp", "kEr4SY8y4ZS8drASBLzLGbQ0Zkm.webp", "tq3h43fZy0H80vzf47MAY7R9Mxo.webp", "ucDMmWAYOiBhd6cZCDWVK9DIIQ4.webp"],
    synopsis: "Will est un petit bouc avec de grands rêves. Lorsqu'il décroche une chance inespérée de rejoindre la ligue professionnelle de \"roarball\", un sport ultra-intense réservé aux bêtes les plus rapides et féroces, il entend bien saisir l’opportunité. Problème : ses nouveaux coéquipiers ne sont pas vraiment ravis d'avoir un \"petit\" dans l'équipe. Mais Will est prêt à tout pour changer les règles du jeu.",
    genres: ["Animation", "Comédie", "Familial"],
    directors: ["Tyree Dillihay"],
    stars: ["Caleb McLaughlin", "Gabrielle Union", "Stephen Curry"],
    length: "1h 40m"
}, {
    id: 9,
    title: "Dune",
    src: "yyrBBEHvwdJEgNgALxuh0EyWmsN.webp",
    year: "2026",
    backdrops: ["6qMyijXWuDfrN8bXUJT1DErjQog.webp", "imkZLxabgZBdsWPnkPkFCHC9NCw.webp", "kEr4SY8y4ZS8drASBLzLGbQ0Zkm.webp", "tq3h43fZy0H80vzf47MAY7R9Mxo.webp", "ucDMmWAYOiBhd6cZCDWVK9DIIQ4.webp"],
    synopsis: "L'histoire de Paul Atreides, jeune homme aussi doué que brillant, voué à connaître un destin hors du commun qui le dépasse totalement. Car, s'il veut préserver l'avenir de sa famille et de son peuple, il devra se rendre sur Dune, la planète la plus dangereuse de l'Univers. Mais aussi la seule à même de fournir la ressource la plus précieuse capable de décupler la puissance de l'Humanité. Tandis que des forces maléfiques se disputent le contrôle de cette planète, seuls ceux qui parviennent à dominer leur peur pourront survivre…",
    genres: ["Science-Fiction", "Aventure"],
    directors: ["Denis Villeneuve"],
    stars: ["Timothée Chalamet", "Rebecca Ferguson", "Oscar Isaac", "Jason Momoa"],
    length: "2h 35m"
},];
