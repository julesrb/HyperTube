"use client";

import {movies} from "@/types/movie";
import React from "react";
import MovieCard from "@/components/MovieCard";

export default function HomePage() {
    return (<div className="">
            <div className="grid grid-cols-3 gap-4 px-4">
                {movies.map((movie, index) => (<MovieCard key={index} movie={movie}/>))}
            </div>
            <p style={{fontFamily: 'FT Calhern', fontWeight: 900, fontStyle: 'normal'}}>HyperTube | FT Calhern</p>
            <p style={{fontFamily: 'FT Calhern', fontWeight: 'bold', fontStyle: 'normal'}}>HyperTube | FT Calhern</p>
            <p style={{fontFamily: 'FT Calhern Hairline', fontWeight: 'normal', fontStyle: 'normal'}}>HyperTube | FT
                Calhern Hairline</p>
            <p style={{fontFamily: 'FT Calhern', fontWeight: 900, fontStyle: 'normal'}}>HyperTube | FT Calhern</p>
            <p style={{fontFamily: 'FT Calhern', fontWeight: 300, fontStyle: 'normal'}}>HyperTube | FT Calhern</p>
            <p style={{fontFamily: 'FT Calhern', fontWeight: 500, fontStyle: 'normal'}}>HyperTube | FT Calhern</p>
            <p style={{fontFamily: 'FT Calhern', fontWeight: 'normal', fontStyle: 'normal'}}>HyperTube | FT Calhern</p>
            <p style={{fontFamily: 'FT Calhern', fontWeight: 600, fontStyle: 'normal'}}>HyperTube | FT Calhern</p>
            <p style={{fontFamily: 'FT Calhern', fontWeight: 100, fontStyle: 'normal'}}>HyperTube | FT Calhern</p>
            <p style={{fontFamily: 'FT Calhern', fontWeight: 200, fontStyle: 'normal'}}>HyperTube | FT Calhern</p>
            <p style={{fontFamily: 'FT Calhern Ultrathin', fontWeight: 100, fontStyle: 'normal'}}>HyperTube | FT Calhern
                Ultrathin</p>
            <p style={{fontFamily: 'FT Calhern Wide', fontWeight: 900, fontStyle: 'normal'}}>HyperTube | FT Calhern
                Wide</p>
            <p style={{fontFamily: 'FT Calhern Wide', fontWeight: 900, fontStyle: 'normal'}}>HyperTube | FT Calhern
                Wide</p>
            <p style={{fontFamily: 'FT Calhern Wide', fontWeight: 'bold', fontStyle: 'normal'}}>HyperTube | FT Calhern
                Wide</p>
            <p style={{fontFamily: 'FT Calhern Wide Hairline', fontWeight: 'normal', fontStyle: 'normal'}}>HyperTube |
                FT Calhern Wide Hairline</p>
            <p style={{fontFamily: 'FT Calhern Wide', fontWeight: 900, fontStyle: 'normal'}}>HyperTube | FT Calhern
                Wide</p>
            <p style={{fontFamily: 'FT Calhern Wide', fontWeight: 300, fontStyle: 'normal'}}>HyperTube | FT Calhern
                Wide</p>
            <p style={{fontFamily: 'FT Calhern Wide', fontWeight: 500, fontStyle: 'normal'}}>HyperTube | FT Calhern
                Wide</p>
            <p style={{fontFamily: 'FT Calhern Wide', fontWeight: 'normal', fontStyle: 'normal'}}>HyperTube | FT Calhern
                Wide</p>
            <p style={{fontFamily: 'FT Calhern Wide', fontWeight: 600, fontStyle: 'normal'}}>HyperTube | FT Calhern
                Wide</p>
            <p style={{fontFamily: 'FT Calhern Wide', fontWeight: 100, fontStyle: 'normal'}}>HyperTube | FT Calhern
                Wide</p>
            <p style={{fontFamily: 'FT Calhern Wide', fontWeight: 200, fontStyle: 'normal'}}>HyperTube | FT Calhern
                Wide</p>
            <p style={{fontFamily: 'FT Calhern Wide Ultrathin', fontWeight: 100, fontStyle: 'normal'}}>HyperTube | FT
                Calhern Wide Ultrathin</p>
        </div>);
}
