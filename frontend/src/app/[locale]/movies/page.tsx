"use client";

import {iMovie} from "@/types/movie";
import {iGenre} from "@/types/genre";
import React, {useEffect, useRef, useState} from "react";
import {GridIcon, ListIcon} from "@/components/Icons";
import {useSearchParams} from "next/navigation";
import {useLocale, useTranslations} from "next-intl";
import {useGenres} from "@/hooks/useGenres";
import {tLocale} from "@/i18n/request";
import {useMovies} from "@/services/movies.service";
import computeTotalPage from "@/utils/computeTotalPage";
import Pagination from "@/components/ui/Pagination";
import CloseButton from "@/components/ui/Button/CloseButton";
import useModal from "@/contexts/ModalContext";
import {useResponsiveSize} from "@/hooks/useResponsiveSize";
import MoviesCard from "@/components/features/movie/MoviesCard";
import MovieCardList from "@/components/features/movie/MovieCardList";

type tViewType = | "grid" | "list";
type tSort = "title" | "genre" | "grade" | "year";

interface iSort {
    type?: tSort;
    side: boolean;
}

export default function Page() {
    const searchParams = useSearchParams();
    const genreId = searchParams.get("genre") as number | null;
    let genre: undefined | iGenre;
    const locale = useLocale() as tLocale;
    const {data} = useGenres(locale);
    if (genreId && data)
        genre = data.genres.find(e => e.id == genreId);
    const mostRated = searchParams.get("sort");
    const query = searchParams.get("q");
    const [searchValue, setSearchValue] = useState(query ?? "");
    const [viewType, setViewType] = useState<tViewType>(genre === undefined && mostRated === null ? "grid" : "list");
    const [sort, setSort] = useState<iSort>({type: mostRated ? "grade" : undefined, side: true});
    const [index, setIndex] = useState(0);
    const {data: movies} = useMovies(searchValue.trim(), index);
    const totalPage = computeTotalPage(movies);

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setViewType(genre === undefined && mostRated === null ? "grid" : "list");
    }, [genre, mostRated]);

    const handleSearchChange = (e?: React.ChangeEvent<HTMLInputElement>) => {
        const newValue = e?.target.value.toLowerCase() ?? "";
        setSearchValue(newValue);
    }
    const handleSetViewType = (value: tViewType) => setViewType(value);
    const changeSort = (type: tSort, side: boolean) => setSort({type, side});
    const changeIndex = (newIndex: number) => setIndex(newIndex);

    return (<div className="flex flex-col gap-4 mx-2 md:mx-4 xl:mx-6">
        <SearchBar searchValue={searchValue} onChange={handleSearchChange} />
        <Filter viewType={viewType} onClick={handleSetViewType}/>
        <Pagination currenIndex={index} totalPage={totalPage} onClick={changeIndex} >
            <Results movies={movies?.data} viewType={viewType} sort={sort} changeSort={changeSort} genre={genre}/>
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
        <input ref={inputRef} type="search" placeholder={t("searchPlaceholder")} value={searchValue} onChange={onChange}
        className="w-full bg-white text-5xl md:text-6xl xl:text-8xl font-condensed uppercase border-b focus:border-b-2"></input>
        <CloseButton className="absolute right-10" onClickAction={() => onChange()} disabled={searchValue.length === 0}/>
    </div>);
}

function Filter({viewType, onClick}: {viewType: tViewType, onClick: (value: tViewType) => void}) {
    return (<div className="flex w-full justify-end gap-4 px-6">
        <button onClick={() => onClick("grid")}><GridIcon color={viewType == "grid" ? "black" : "gray"}/></button>
        <button onClick={() => onClick("list")}><ListIcon color={viewType == "list" ? "black" : "gray"}/></button>
    </div>);
}

function Results({movies, viewType, sort, changeSort, genre}: {movies?: iMovie[], viewType: tViewType, sort: iSort, changeSort: (type: tSort, side: boolean) => void, genre: undefined | iGenre}) {
    const {openModal} = useModal();
    const [filterGenre, setFilterGenre] = useState<iGenre[]>(genre === undefined ? [] : [genre])
    const size = useResponsiveSize();
    const t = useTranslations("movies");

    const noResult = () => (<p className="small-text">{t("noResults")}</p>);

    if (movies && movies.length === 0)
        return noResult();

    if (viewType === "grid")
        return (<MoviesCard movieSets={movies}/>);

    const sortOptions: {type: tSort, label: string}[] = [
        {type: "title", label: t("sort.title")},
        {type: "year", label: t("sort.year")},
        {type: "genre", label: t("sort.genre")},
        {type: "grade", label: t("sort.rating")},
    ];
    let sortedMovies = movies;
    if (sortedMovies) {
        if (sort.type === "grade")
            sortedMovies = sortedMovies.sort((a, b) => a.note - b.note);
        else if (sort.type === "year")
            sortedMovies = sortedMovies.sort((a, b) => parseInt(a.year) - parseInt(b.year));
        else if (sort.type === "title")
            sortedMovies = sortedMovies.sort((a, b) => b.title.localeCompare(a.title));

        if (sort.side)
            sortedMovies = sortedMovies.reverse();

        if (filterGenre.length > 0 && size === "xl")
            sortedMovies = sortedMovies.filter(m => {
                for (let i = 0; i < filterGenre.length; i++) {
                    if (m.genres && !m.genres.includes(filterGenre[i].id))
                        return false;
                }
                return true;
            })
    }


    const handleSort = (sortOption: tSort) => {
        if (sortOption === "genre")
            openModal({type: "filter-genre", filterGenre: [filterGenre, setFilterGenre]})
        else
            changeSort(sortOption, sort.type === sortOption ? !sort.side : true)
    }

    const deleteGenre = (genre: iGenre[]) => {
        let newGenre = filterGenre.filter(g => !genre.find(deletedGenre => deletedGenre.id === g.id));
        if (newGenre.length === filterGenre.length)
            newGenre = filterGenre.slice(0, 2);
        setFilterGenre(newGenre);
    }

    const classNames = ["sm:pl-3", "", "hidden lg:table-cell", "hidden sm:table-cell"]

    return (<div>
        <table className="table-fixed w-full overflow-hidden">
            <colgroup>
                <col className="w-30 sm:w-55 xl:w-80" />
                <col />
                <col className="w-0" />
                <col className="w-1/4 hidden lg:table-column" />
                <col className="w-15 hidden sm:table-column" />
                <col className="w-32" />
            </colgroup>

            <thead>
                <tr className="text-left align-top">
                    <th></th>
                    {sortOptions.map((sortOption, i) =>
                        <th key={sortOption.type} className={classNames[i]}>
                            <button className={"relative capitalize text-nowrap hover:underline text-xs sm:text-base" + (sortOption.type === "year" ? " -left-4 sm:-left-20 md:-left-30 xl:-left-45 2xl:-left-80" : "")}
                                    onClick={() => handleSort(sortOption.type)}>
                                {sortOption.label} {sortOption.type === sort.type && (sort.side ? "▾" : "▴")}
                            </button>
                            {sortOption.type === "genre" && <SelectedGenre genres={filterGenre} deleteGenre={deleteGenre}/>}
                        </th>
                    )}
                    <th></th>
                </tr>
            </thead>
            <tbody>
                {sortedMovies ?
                    sortedMovies.map((movie) => (<MovieCardList key={movie.imdb_id} movie={movie} setFilterGenre={setFilterGenre}/>)) :
                    [...Array(6)].map((_, i) => (<MovieCardList key={i} movie={null} setFilterGenre={setFilterGenre}/>))
                }
            </tbody>
        </table>
        {(sortedMovies && sortedMovies.length === 0) && noResult()}
    </div>);
}

function SelectedGenre({genres, deleteGenre}: {genres: iGenre[], deleteGenre:(genre: iGenre[]) => void}) {
    const showGenres = genres.slice(0, 2);
    const t = useTranslations("movies");

    return (<div className="flex gap-2">
        {showGenres.map((genre, index) => (<div key={index}
        className="border flex items-center">
            <span className="font-hairline tracking-wider text-sm px-2 text-nowrap">{genre.name}</span>
            <CloseButton size={20} className="border-l px-1" onClickAction={() => deleteGenre([genre])} />
        </div>))}
        { genres.length > 2 && <div className="border flex items-center">
            <span className="font-hairline tracking-wider text-sm px-2 text-nowrap">{t("selectedGenres.more", {count: genres.length - 2})}</span>
            <CloseButton size={20} className="border-l px-1" onClickAction={() => deleteGenre(genres.slice(2))} />
        </div>}
    </div>);
}
