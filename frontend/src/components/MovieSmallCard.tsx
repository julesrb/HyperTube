import Image from "next/image";
import {tMovie} from "@/types/movie";
import {useTranslations} from "next-intl";


export default function MovieSmallCard({movie} : {movie: tMovie}) {
    const t = useTranslations("movie");
    return (
        <div className="relative w-48 h-72 overflow-hidden cursor-pointer group border">
            <Image className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105" width={100} height={100} src={"/images/" + movie.src} alt={t("posterAlt", {title: movie.title})}/>
            <div className="absolute inset-0 bg-linear-to-t from-black via-transparent to-transparent opacity-0 group-hover:opacity-80 transition-opacity duration-300 flex items-end p-4">
                <h3 className="text-white text-lg font-bold">{movie.title}</h3>
            </div>
        </div>);
}
