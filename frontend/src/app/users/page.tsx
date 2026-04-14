"use client";

import Input from "@/components/Input";
import React, {useState} from "react";
import Button from "@/components/Button";
import DefaultUserIcon from "@/components/DefaultUserIcon";
import {movies} from "@/types/movie";
import MovieCard from "@/components/MovieCard";

export default function Page() {
    const tabs = {profile: ProfileTab, avatar: AvatarTab, auth: AuthTab, history: MovieHistoryTab};
    const [activeTab, setActiveTab] = useState<keyof typeof tabs>("profile");
    const ActiveTab = tabs[activeTab];

    const switchTab = (tabName: keyof typeof tabs) => {
        if (activeTab !== tabName) {
            setActiveTab(tabName);
        }
    }

    return (<div className="px-4">
        <div className="flex items-center gap-4 justify-center mt-10 mb-16">
            <h3 className="font-bold rounded-full bg-purple size-22 flex items-center justify-center border">
                FG
            </h3>
            <div className="flex flex-col items-start">
                <h2>Florian G.</h2>
                <p className="uppercase">Member since 10.04.2026</p>
            </div>
        </div>
        <div className="flex h-16 my-10">
            <div className="border-b border-r w-12"></div>
            {(Object.keys(tabs) as Array<keyof typeof tabs>).map((tabName, index) => (<button
                key={index}
                className={"border-t border-r px-16 border-b" + (activeTab === tabName ? " border-b-white" : "")}
                onClick={() => switchTab(tabName)}><h4>{tabName}</h4></button>))}
            <div className="border-b w-full"></div>
        </div>
        <div>
            <ActiveTab/>
        </div>

    </div>);
}


function ProfileTab() {
    return (<div className="flex flex-col gap-4 items-start max-w-2/5 mx-auto">
        <Input type="email" placeholder="Email"></Input>

        <div className="flex gap-2 w-full">
            <Input type="firstname" placeholder="Firstname"></Input>
            <Input type="lastname" placeholder="Lastname"></Input>
        </div>

        <Input type="username" placeholder="Username" className={"max-w-3/5"}></Input>

        <Button className="h-8">Save Changes</Button>
    </div>);
}


function AvatarTab() {
    return (<div className="flex flex-col gap-2 items-center justify-center">
        <DefaultUserIcon className="mb-6"/>
        <Button>Select New avatar</Button>

        {/* TODO use SmallButton*/}
        <button
            className={(true ? "text-red  custom-h-underline-red" : "text-gray") + " text-sm font-sans"}>Remove
        </button>
    </div>);
}



function AuthTab() {
    return (<div className="max-w-2/5 mx-auto flex flex-col items-start gap-4">

        <Input type="password" placeholder="Current password"></Input>
        <Input type="password" placeholder="New password"></Input>
        <Input type="password" placeholder="Confirm new password"></Input>

        <Button>Change</Button>
    </div>);
}


function MovieHistoryTab() {
    return (<div className="grid grid-cols-3 gap-4">
        {movies.map((movie, index) => (<MovieCard key={index} movie={movie}/>))}
    </div>);
}
