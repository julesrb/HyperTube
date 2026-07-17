import {iTorrent} from "@/types/movie";

export default function getBestTorrent(torrents?: iTorrent[]) {
    if (!torrents || !torrents.length)
        return null;

    const qualityScore: Record<string, number> = {
        "2160p": 100,
        "1080p": 70,
        "720p": 40,
        "480p": 20,
    };

    const languageScore: Record<string, number> = {
        "MULTI": 30,
        "TRUEFRENCH": 25,
        "FRENCH": 20,
        "VOSTFR": 10,
        "VO": 0,
    };

    function getExpectedSize(quality: string) {
        switch (quality) {
            case "2160p":
                return { min: 8, ideal: 20, max: 80 };

            case "1080p":
                return { min: 4, ideal: 12, max: 35 };

            case "720p":
                return { min: 2, ideal: 5, max: 15 };

            default:
                return { min: 1, ideal: 3, max: 10 };
        }
    }

    function sizeScore(size: number, quality: string) {
        const { min, ideal, max } = getExpectedSize(quality);

        if (size < min)
            return -20;

        if (size > max)
            return -5;

        const distance = Math.abs(size - ideal);
        return 30 - distance;
    }

    const ranked = torrents.map((torrent) => {
        const seeds = parseInt(torrent.seeds, 10);
        const seedScore = Math.log10(seeds + 1) * 30;

        const score =
            (qualityScore[torrent.quality] || 0) +
            (languageScore[torrent.language] || 0) +
            seedScore +
            sizeScore(torrent.size, torrent.quality);

        return {
            torrent,
            score,
        };
    });

    ranked.sort((a, b) => b.score - a.score);

    return ranked[0].torrent;
}
