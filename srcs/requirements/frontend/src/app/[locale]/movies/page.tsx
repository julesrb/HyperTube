"use client";

import {iMovie} from "@/types/movie";
import {iGenre} from "@/types/genre";
import React, {useEffect, useRef, useState} from "react";
import {GridIcon, ListIcon} from "@/components/Icons";
import {useLocale, useTranslations} from "next-intl";
import useGenres from "@/hooks/useGenres";
import {tLocale} from "@/i18n/request";
import {useMovies} from "@/services/movies.service";
import computeTotalPage from "@/utils/computeTotalPage";
import Pagination from "@/components/ui/Pagination";
import CloseButton from "@/components/ui/Button/CloseButton";
import MoviesGrid from "@/components/features/movie/MoviesGrid";
import MoviesList from "@/components/features/movie/MoviesList";
import SmallText from "@/components/ui/SmallText";
import {useUserFilmHistory} from "@/services/users.service";
import useAuth from "@/contexts/AuthContext";
import addProgressToMovie from "@/utils/addProgressToMovie";
import IconButton from "@/components/ui/Button/IconButton";
import {usePathname, useRouter} from "@/i18n/navigation";
import {useSearchParams} from "next/navigation";

type tViewType = | "grid" | "list";

export type tSort = "title" | "genre" | "grade" | "year";
export interface iSort {
    type?: tSort;
    side: boolean;
}

export default function Page() {
    const searchParams = useSearchParams();
    const router = useRouter();
    const pathname = usePathname();
    const genreId = searchParams.get("genre") as number | null;
    let genre: undefined | iGenre;
    const locale = useLocale() as tLocale;
    const {data} = useGenres(locale);
    if (genreId && data)
        genre = data.genres.find(e => e.id == genreId);
    const mostRated = searchParams.get("sort");
    const query = searchParams.get("q");
    const [searchValue, setSearchValue] = useState(query ?? "");
    const [viewType, setViewType] = useState<tViewType>("grid");
    const [sort, setSort] = useState<iSort>({type: mostRated ? "grade" : undefined, side: true});
    const [index, setIndex] = useState(0);
    const {data: rawMovies} = useMovies(searchValue.trim(), index);
    const {user} = useAuth();
    const {data: watchHistory} = useUserFilmHistory(user?.id);
    const movies = addProgressToMovie(watchHistory?.data, rawMovies?.data) as iMovie[] | undefined;
    const totalPage = computeTotalPage(rawMovies);

    useEffect(() => {
        const savedViewType = localStorage.getItem("searchViewType") as tViewType;
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setViewType((genre === undefined && mostRated === null ? (savedViewType ?? "grid") : "list"));
    }, [genre, mostRated]);

    useEffect(() => {
        const q = searchValue.trim();
        const params = new URLSearchParams(searchParams.toString());
        if (q)
            params.set("q", q);
        else
            params.delete("q");
        router.push(`${pathname}?${params.toString()}`);
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [rawMovies]);

    const handleSearchChange = (e?: React.ChangeEvent<HTMLInputElement>) => {
        const newValue = e?.target.value.toLowerCase() ?? "";
        setSearchValue(newValue);
    }
    const handleSetViewType = (value: tViewType) => {
        localStorage.setItem("searchViewType", value);
        setViewType(value);
    };
    const changeSort = (type: tSort, side: boolean) => setSort({type, side});
    const changeIndex = (newIndex: number) => setIndex(newIndex);

    return (<div className="flex flex-col gap-2 md:gap-4 mx-2 md:mx-4 xl:mx-6 pb-2 md:pb-4 xl:pb-6">
        <SearchBar searchValue={searchValue} onChange={handleSearchChange} />
        <Filter viewType={viewType} onClick={handleSetViewType}/>
        <Pagination currenIndex={index} totalPage={totalPage} onClick={changeIndex} variableMT={true}>
            <Results movies={movies} viewType={viewType} sort={sort} changeSort={changeSort} genre={genre}/>
        </Pagination>
    </div>);
}

function SearchBar({searchValue, onChange}: {searchValue: string, onChange: (e?: React.ChangeEvent<HTMLInputElement>) => void}) {
    const inputRef = useRef<HTMLInputElement>(null);
    const t = useTranslations("movies");
    useEffect(() => {
        const el = inputRef.current;
        if (!el) return;
        el.focus();
        el.setSelectionRange(el.value.length, el.value.length);
    }, []);

    return (<div className="flex items-center px-6">
        <input id="search-bar" ref={inputRef} type="search" placeholder={t("searchPlaceholder")} value={searchValue} onChange={onChange}
        className="w-full bg-white text-5xl md:text-7xl xl:text-8xl 2xl:text-9xl font-condensed uppercase border-b focus:border-b-2" />
        <CloseButton className="absolute right-10" onClickAction={() => onChange()} disabled={searchValue.length === 0}/>
    </div>);
}

function Filter({viewType, onClick}: {viewType: tViewType, onClick: (value: tViewType) => void}) {
    return (<div className="flex w-full justify-end gap-4 px-6">
        <IconButton color={viewType == "grid" ? "black" : "gray"} onClick={() => onClick("grid")}>{(color: string) => <GridIcon color={color}/>}</IconButton>
        <IconButton color={viewType == "list" ? "black" : "gray"} onClick={() => onClick("list")}>{(color: string) => <ListIcon color={color}/>}</IconButton>
    </div>);
}

function Results({movies, viewType, sort, changeSort, genre}: {movies?: iMovie[], viewType: tViewType, sort: iSort, changeSort: (type: tSort, side: boolean) => void, genre: undefined | iGenre}) {
    const t = useTranslations("movies");

    if (movies && movies.length === 0)
        return (<SmallText>{t("noResults")}</SmallText>);

    if (viewType === "grid")
        return (<MoviesGrid movieSets={movies}/>);
    return (<MoviesList movieSets={movies} sort={sort} changeSort={changeSort} genre={genre}/>);
}
