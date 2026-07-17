import {useState} from "react";
import {useTranslations} from "next-intl";
import computeTotalPage from "@/utils/computeTotalPage";
import Pagination from "@/components/ui/Pagination";
import MoviesGrid from "@/components/features/movie/MoviesGrid";
import {useUserFilmHistory} from "@/services/users.service";
import {iUser} from "@/types/user";
import SmallText from "@/components/ui/SmallText";

export default function MovieHistoryTab({user}: {user: iUser}) {
    const [index, setIndex] = useState(0);
    const changeIndex = (newIndex: number) => {setIndex(newIndex);}
    const t = useTranslations("profile");
    const {data: watchMovies} = useUserFilmHistory(user.id);
    const totalPage = computeTotalPage(watchMovies);

    if (!watchMovies || watchMovies.data.length === 0)
        return (<SmallText>{t("noMoviesYet")}</SmallText>);
    return (<Pagination currenIndex={index} onClick={changeIndex} totalPage={totalPage} variableMT={true}>
        <MoviesGrid movieSets={watchMovies.data}/>
    </Pagination>);
}
