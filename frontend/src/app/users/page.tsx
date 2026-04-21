"use client";

import React, {useState} from "react";
import {movies} from "@/types/movie";
import {MovieCard} from "@/components/MovieCard";
import ProfilePicture from "@/components/ProfilePicture";
import {tUser} from "@/types/user";
import ProfileTab from "@/app/users/ProfileTab";
import AuthTab from "@/app/users/AuthTab";
import {useAuth} from "@/context/AuthContext";

export default function Page() {
    const {user} = useAuth();
    if (!user)
        return null;
    const [currentUser, setUser] = useState(user);
    const tabs = {profile: ProfileTab, auth: AuthTab, history: MovieHistoryTab};
    const [activeTab, setActiveTab] = useState<keyof typeof tabs>("profile");
    const ActiveTab = tabs[activeTab];
    const date = new Date(currentUser.joined_at);

    const switchTab = (tabName: keyof typeof tabs) => {
        if (activeTab !== tabName)
            setActiveTab(tabName);
    }

    return (<div className="px-4">
        <div className="flex items-center gap-4 justify-center mt-10 mb-16">
            <ProfilePicture user={currentUser} size={1}/>
            <div className="flex flex-col items-start">
                <h2>{currentUser.firstname} {currentUser.lastname[0]}.</h2>
                <p className="uppercase">Member since {date.toLocaleDateString('fr-FR').replace(/\//g, '.')}</p>
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
        <ActiveTab user={currentUser} setUser={setUser}/>
    </div>);
}

function MovieHistoryTab({user}: {user: tUser}) {
    if (user.film_history.length === 0)
        return (<div className="flex justify-center pt-5"><p>You haven&#39;t seen any films yet.</p></div>);
    return (<div className="grid grid-cols-3 gap-4">
        {user.film_history.map((movieIdx, index) => (<MovieCard key={index} movie={movies[movieIdx]}/>))}
    </div>);
}
