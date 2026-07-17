import {iMovie, iMovieDetails} from "@/types/movie";
import {iCommentDetails} from "@/types/comment";

    export default function addProgressToMovie(history?: iMovie[], items?: iMovie[] | iCommentDetails[] | iMovieDetails) {
    if (!history || !items || ("length" in items && items.length <= 0))
        return items;
    else if ("tmdb_id" in items) {
        const historyMovie = history.find(m => m.imdb_id === items.imdb_id);
        if (historyMovie) {
            items.progress = historyMovie.progress;
            items.complete = historyMovie.complete;
        }
        return items;
    } else if ("content" in items[0]) {
        const comments = items as iCommentDetails[];
        return comments.map((comment) => {
            const historyMovie = history.find(m => m.imdb_id === comment.movie.imdb_id);
            if (historyMovie) {
                comment.movie.progress = historyMovie.progress;
                comment.movie.complete = historyMovie.complete;
            }
            return comment;
        })
    } else if ("imdb_id" in items[0]) {
        const movies = items as iMovie[];
        return movies.map((movie) => {
            const historyMovie = history.find(m => m.imdb_id === movie.imdb_id);
            if (historyMovie) {
                movie.progress = historyMovie.progress;
                movie.complete = historyMovie.complete;
            }
            return movie;
        })
    }
}
