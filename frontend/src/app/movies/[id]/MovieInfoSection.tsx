import {tmovie} from "@/types/movie";
import GenreTag from "@/components/GenreTag";

export default function MovieInfoSection({movie} : {movie: tmovie}) {
    return (<div className="flex flex-col mt-3 gap-4 max-w-2/3 mx-auto">
        <h1 className="relative mx-auto mb-6 text-center">{movie.title}
        <span className="absolute -right-18 font-hairline text-3xl tracking-normal">{movie.year}</span>
    </h1>
    <InfoMovie name="Length">
        <p>{movie.length}</p>
    </InfoMovie>

    <InfoMovie name="Genre">
        <div className="flex gap-4">
            {movie.genres.map((genre) => (<GenreTag key={genre}>{genre}</GenreTag>))}
        </div>
    </InfoMovie>

    <InfoPeoplesMovie name="Directeor" items={movie.directors}/>

    <InfoPeoplesMovie name="Stars" items={movie.stars}/>

    <InfoMovie name="Synopsis">
        <p>{movie.synopsis}</p>
    </InfoMovie>

</div>);
}


function InfoMovie({children, name}: { children: React.ReactNode, name: string }) {
    return (<div className="flex gap-4">
        <div className={"flex justify-end w-1/2" + (name === "Synopsis" ? "" : " items-center")}>
            <span className="font-bold">{name}</span>
        </div>
        <div className="w-1/2">
            {children}
        </div>
    </div>);
}

function InfoPeoplesMovie({name, items}: { name: string, items: string[] }) {
    return (<InfoMovie name={name}>
        <p>
            {items.map((i, index) => (<span key={index}>
                    <span className="custom-underline hover:cursor-pointer">{i}</span>
                {index < items.length - 1 && " ,   "}
                </span>))}
        </p>
    </InfoMovie>);
}
