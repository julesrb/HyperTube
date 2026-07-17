import {iMovie} from "@/types/movie";

export default function shuffleArray(array?: iMovie[]) {
    if (!array)
        return []

    const newArray = [...array];

    for (let i = newArray.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));

        [newArray[i], newArray[j]] = [newArray[j], newArray[i]];
    }

    return newArray;
}