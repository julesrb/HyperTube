import {iUser} from "@/types/user";
import {iMovie} from "@/types/movie";

export default function WatchProgress({user, movie}: {user?: iUser, movie?: iMovie}) {
    if (user && movie?.pourcent && movie.pourcent > 0)
        return (<div className={`absolute bottom-0 h-1 bg-${user.color} z-10`} style={{width: `${movie.pourcent}%`}} />);
}
