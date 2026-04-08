import Image from "next/image";
import {tmovies} from "@/types/movie";


export default function MoviesSmallCard({movie} : {movie: tmovies}) {
    return (
        <div className="relative w-48 h-72 overflow-hidden cursor-pointer group custom-border rounded-sm">
            <Image className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105" width={100} height={100} src={"/images/" + movie.src} alt={"poster of movie: " + movie.title}/>
            {/*<Image className="rounded-sm hover:custom-border" width={100} height={100} src={"/images/" + film.src} alt={"poster of film: " + film.name}/>*/}
            <div className="absolute inset-0 bg-gradient-to-t from-black via-transparent to-transparent opacity-0 group-hover:opacity-80 transition-opacity duration-300 flex items-end p-4">
                <h3 className="text-white text-lg font-bold">{movie.title}</h3>
            </div>
        </div>);
}
