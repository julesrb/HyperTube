"use client";

import {movies} from "@/types/movie";
import {ListMovieCard, MoviesCard} from "@/components/MovieCard";
import React, {useEffect, useRef, useState} from "react";
import {GridIcon, ListIcon} from "@/components/Icons";
import {CloseButton} from "../../components/Buttons";
import {useModal} from "@/context/ModalContext";
import {useSearchParams} from "next/navigation";
import Pagination from "@/components/Pagination";

type tViewType = | "grid" | "list";
type tSort = "name" | "genre" | "grade" | "year";

interface iSort {
    type: tSort;
    side: boolean;
}

export default function Page() {
    const searchParams = useSearchParams();
    const genre = searchParams.get('genre');
    const mostRated = searchParams.get('sort');
    const [searchValue, setSearchValue] = useState("");
    const [viewType, setViewType] = useState<tViewType>(genre === null && mostRated === null ? "grid" : "list");
    const [sort, setSort] = useState<iSort>({type: mostRated ? "grade" : "name", side: true});
    const [index, setIndex] = useState(0);

    const handleSearchChange = (e?: React.ChangeEvent<HTMLInputElement>) => {
        const newValue = e === undefined ? "" : e.target.value.toLowerCase()
        setSearchValue(newValue);
    }
    const handleSetViewType = (value: tViewType) => { setViewType(value); }
    const changeSort = (type: tSort, side: boolean) => { setSort({type, side}); }
    const changeIndex = (newIndex: number) => {setIndex(newIndex);}

    return (<div className="flex flex-col gap-4 mx-2 md:mx-4 xl:mx-6">
        <SearchBar searchValue={searchValue} onChange={handleSearchChange} />
        <Filter viewType={viewType} onClick={handleSetViewType}/>
        <Pagination currenIndex={index} totalPage={5} onClick={changeIndex} >
            <Results searchValue={searchValue} viewType={viewType} sort={sort} changeSort={changeSort} genre={genre}/>
        </Pagination>
    </div>);
}

function SearchBar({searchValue, onChange}: {searchValue: string, onChange: (e?: React.ChangeEvent<HTMLInputElement>) => void}) {
    const inputRef = useRef<HTMLInputElement>(null);
    useEffect(() => {
        const el = inputRef.current;
        if (!el) return;
        el.focus();
    }, []);

    return (<div className="flex items-center px-6">
        <input ref={inputRef} type="search" placeholder="Rechercher un film..." value={searchValue} onChange={onChange}
        className="w-full bg-white text-5xl md:text-7xl xl:text-9xl font-condensed uppercase border-b focus:border-b-2"></input>
        <CloseButton className="absolute right-10" onClick={() => onChange()} disabled={searchValue.length === 0}/>
    </div>);
}

function Filter({viewType, onClick}: {viewType: tViewType, onClick: (value: tViewType) => void}) {
    return (<div className="flex w-full justify-end gap-4 px-6">
        <button onClick={() => onClick("grid")}><GridIcon color={viewType == "grid" ? "black" : "gray"}/></button>
        <button onClick={() => onClick("list")}><ListIcon color={viewType == "list" ? "black" : "gray"}/></button>
    </div>);
}

function Results({searchValue, viewType, sort, changeSort, genre}: {searchValue: string, viewType: tViewType, sort: iSort, changeSort: (type: tSort, side: boolean) => void, genre: null | string}) {
    const {openModal} = useModal();
    const [filterGenre, setFilterGenre] = useState<string[]>(genre === null ? [] : [genre])
    const filteredMovies = movies.filter((movie) => movie.title.toLowerCase().includes(searchValue.trim()));

    if (filteredMovies.length === 0)
        return (<p>Aucun film trouvé</p>);

    if (viewType === "grid")
        return (<MoviesCard movieSets={filteredMovies}/>);

    const sortOptions: tSort[] = ["name", "year", "genre", "grade"];
    let sortedMovies;
    if (sort.type === "grade")
        sortedMovies = filteredMovies.sort((a, b) => a.rate - b.rate);
    else if (sort.type === "year")
        sortedMovies = filteredMovies.sort((a, b) => parseInt(a.year) - parseInt(b.year));
    else if (sort.type === "genre")
        sortedMovies = filteredMovies.sort((a, b) => a.id - b.id);
    else
        sortedMovies = filteredMovies.sort((a, b) => b.title.localeCompare(a.title));

    if (sort.side)
        sortedMovies = sortedMovies.reverse();

    if (filterGenre.length > 0)
        sortedMovies = sortedMovies.filter(m => {
            for (let i = 0; i < filterGenre.length; i++) {
                if (!m.genres.includes(filterGenre[i]))
                    return false;
            }
            return true;
        })

    const handleSort = (sortOption: tSort) => {
        if (sortOption === "genre")
            openModal({type: "filter-genre", filterGenre: [filterGenre, setFilterGenre]})
        else
            changeSort(sortOption, sort.type === sortOption ? !sort.side : true)
    }

    const deleteGenre = (genre: string) => {
        let newGenre = filterGenre.filter(g => g !== genre);
        if (newGenre.length === filterGenre.length)
            newGenre = filterGenre.slice(0, 2);
        setFilterGenre(newGenre);
    }

    const classNames = ["sm:pl-3", "", "hidden lg:table-cell", "hidden sm:table-cell"]

    return (<table className="table-fixed w-full overflow-hidden">
        <colgroup>
            <col className="w-30 sm:w-55 xl:w-80" />
            <col />
            <col className="w-0" />
            <col className="w-1/4 hidden lg:table-column" />
            <col className="w-15 hidden sm:table-column" />
            <col className="w-25" />
        </colgroup>

        <thead>
            <tr className="text-left align-top">
                <th></th>
                {sortOptions.map((sortOption, i) =>
                    <th key={sortOption} className={classNames[i]}>
                        <button className={"relative capitalize text-nowrap hover:underline text-xs sm:text-base" + (sortOption === "year" ? " -left-5 md:-left-30 xl:-left-45" : "")}
                                onClick={() => handleSort(sortOption)}>
                            {sortOption} {sortOption === sort.type && (sort.side ? "▾" : "▴")}
                        </button>
                        {sortOption === "genre" && <SelectedGenre genres={filterGenre} deleteGenre={deleteGenre}/>}
                    </th>
                )}
                <th></th>
            </tr>
        </thead>
        <tbody>
            {sortedMovies.map((movie, index) => (<ListMovieCard key={index} movie={movie} setFilterGenre={setFilterGenre}/>))}
        </tbody>
    </table>);
}

function SelectedGenre({genres, deleteGenre}: {genres: string[], deleteGenre:(genre: string) => void}) {
    const showGenres = genres.slice(0, 2);
    if (genres.length > 2)
        showGenres.push(`+ ${genres.length - 2} de plus`)
    return (<div className="flex gap-2">
        {showGenres.map((genre, index) => (<div key={index}
        className="border flex items-center">
            <span className="font-hairline tracking-wider text-sm px-2 text-nowrap">{genre}</span>
            <CloseButton size={20} className="border-l px-1" onClick={() => deleteGenre(genre)} />
        </div>))}
    </div>);
}
