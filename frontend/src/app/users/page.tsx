"use client";

import Input from "@/components/Input";
import React, {useState} from "react";
import Button from "@/components/Button";
import DefaultUserIcon from "@/components/icon/DefaultUserIcon";
import FilmSmallCard from "@/components/MovieSmallCard";
import {movies} from "@/types/movie";

export default function Page() {
    const tabs = {profile: ProfileTab, auth: AuthTab, history: MovieHistoryTab};
    const [activeTab, setActiveTab] = useState<keyof typeof tabs>("profile");
    const ActiveTab = tabs[activeTab];

    const switchTab = (tabName: keyof typeof tabs) => {
        if (activeTab !== tabName) {
            setActiveTab(tabName);
        }
    }

    return (<div className="px-4">
        <div className="flex items-center gap-4 justify-center mt-10 mb-16">
            <span className="rounded-full bg-purple text-black size-22 flex items-center justify-center text-5xl font-condensed border">
                FG
            </span>
            <div className="flex flex-col items-start">
                <span className="text-6xl font-condensed uppercase tracking-wide">Florian G.</span>
                <p className="uppercase">Member since 10.04.2026</p>
            </div>
        </div>
        <div className="flex h-16 my-10">
            <div className="border-b border-r w-12"></div>
            {(Object.keys(tabs) as Array<keyof typeof tabs>).map((tabName, index) => (<button
                key={index}
                className={"uppercase font-condensed tracking-wide text-4xl border-t border-r px-16 border-b" + (activeTab === tabName ? " border-b-white" : "")}
                onClick={() => switchTab(tabName)}>{tabName}</button>))}
            <div className="border-b w-full"></div>
        </div>
        <div>
            <ActiveTab/>
        </div>

    </div>);
}


function ProfileTab() {
    return (<div className="container mx-auto">
        <h3>Profile</h3>
        <div className="grid grid-cols-2">
            <div className="flex flex-col gap-6 items-start max-w-2/3 mx-auto">
                <Input type="email" placeholder="Email"></Input>

                <div className="flex gap-2 w-full">
                    <Input type="firstname" placeholder="Firstname"></Input>
                    <Input type="lastname" placeholder="Lastname"></Input>
                </div>

                <Input type="username" placeholder="Username" className={"max-w-2/3"}></Input>

                <Button className="h-8">Save Changes</Button>
            </div>

            <div className="flex flex-col gap-2 items-center justify-center">
                <DefaultUserIcon className="mb-6"/>
                <Button className="h-8">Select New avatar</Button>

                <button
                    className={(true ? "text-red  custom-h-underline-red" : "text-gray") + " text-sm font-sans"}>Remove
                </button>
            </div>
        </div>
    </div>);
}


function AuthTab() {
    return (<div className="container mx-auto">
        <h3>Password</h3>

        <Input type="password" placeholder="Current password"></Input>
        <Input type="password" placeholder="New password"></Input>
        <Input type="password" placeholder="Confirm new password"></Input>

        <Button className="h-8">Change</Button>
    </div>);
}


function MovieHistoryTab() {
    return (<div className="container mx-auto">
        <h3>Films Watched</h3>
        <div className="grid grid-cols-4">
            {movies.map((movie, index) => (<FilmSmallCard key={index} movie={movie}/>))}
        </div>
    </div>);
}
