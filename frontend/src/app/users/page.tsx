"use client";

import React, {useState} from "react";
import {movies} from "@/types/movie";
import {MoviesCard} from "@/components/MovieCard";
import ProfilePicture from "@/components/ProfilePicture";
import {tUser} from "@/types/user";
import ProfileTab from "@/app/users/ProfileTab";
import AuthTab from "@/app/users/AuthTab";
import {useAuth} from "@/context/AuthContext";

export default function Page() {
    const {user, updateUser} = useAuth();
    const [activeTab, setActiveTab] = useState<keyof typeof tabs>("profile");
    if (!user)
        return null;
    const tabs = {profile: ProfileTab, auth: AuthTab, history: MovieHistoryTab};
    const ActiveTab = tabs[activeTab];
    const date = new Date(user.joined_at);

    const switchTab = (tabName: keyof typeof tabs) => {
        if (activeTab !== tabName)
            setActiveTab(tabName);
    }

    return (<div className="px-4">
        <div className="flex items-center gap-4 justify-center mt-10 mb-16">
            <ProfilePicture user={user} size={1}/>
            <div className="flex flex-col items-start">
                <h2>{user.firstname} {user.lastname[0]}.</h2>
                <p className="uppercase">Member since {date.toLocaleDateString('fr-FR').replace(/\//g, '.')} #{user.username}</p>
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
        <ActiveTab user={user} updateUser={updateUser}/>
    </div>);
}

function MovieHistoryTab({user}: {user: tUser}) {
    if (user.watch_history.length === 0)
        return (<div className="flex justify-center pt-5"><p>You haven&#39;t seen any films yet.</p></div>);
    return (<MoviesCard movies={user.watch_history.map(m => movies[m.movie_id])}/>);
}
