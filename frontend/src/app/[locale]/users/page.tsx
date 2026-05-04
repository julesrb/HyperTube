"use client";

import React, {useState} from "react";
import {movies} from "@/types/movie";
import {MoviesCard} from "@/components/MovieCard";
import ProfilePicture from "@/components/ProfilePicture";
import {tUser} from "@/types/user";
import ProfileTab from "@/app/[locale]/users/ProfileTab";
import AuthTab from "@/app/[locale]/users/AuthTab";
import {Comments} from "@/components/Comments";
import {useAuth} from "@/context/AuthContext";
import {comments} from "@/types/comment";
import Pagination from "@/components/Pagination";
import {useSearchParams} from "next/navigation";

export default function Page() {
    const tabs = {profile: ProfileTab, auth: AuthTab, history: MovieHistoryTab, comments: CommentsTab};
    type tTab = keyof typeof tabs;

    const searchParams = useSearchParams();
    const tabParam = searchParams.get("tab");
    const initialTab: tTab = tabParam && tabParam in tabs ? (tabParam as tTab) : "profile"
    const {user, updateUser} = useAuth();
    const [activeTab, setActiveTab] = useState<tTab>(initialTab);
    if (!user)
        return null;
    const ActiveTab = tabs[activeTab];
    const date = new Date(user.joined_at);

    const switchTab = (tabName: keyof typeof tabs) => {
        if (activeTab !== tabName)
            setActiveTab(tabName);
    }
    return (<div className="flex flex-col gap-6 sm:gap-10 xl:gap-17 px-2 sm:px-4">
        <div></div>
        <div className="flex items-center gap-4 justify-center">
            <ProfilePicture user={user} size={1}/>
            <div className="flex flex-col items-start">
                <h2>{user.firstname} {user.lastname[0]}.</h2>
                <p className="uppercase">Member since {date.toLocaleDateString('fr-FR').replace(/\//g, '.')} · #{user.username}</p>
            </div>
        </div>
        <div className="flex h-12 sm:h-16">
            <div className="border-b border-r w-12"></div>
            {(Object.keys(tabs) as Array<keyof typeof tabs>).map((tabName, index) => (<button
                key={index}
                className={"custom-condensed text-2xl sm:text-4xl tracking-wide sm:tracking-normal border-t border-r px-4 sm:px-12 xl:px-16 border-b" + (activeTab === tabName ? " border-b-white" : "")}
                onClick={() => switchTab(tabName)}>{tabName}</button>))}
            <div className="border-b w-full"></div>
        </div>
        <ActiveTab user={user} updateUser={updateUser}/>
    </div>);
}

function MovieHistoryTab({user}: {user: tUser}) {
    const [index, setIndex] = useState(0);
    const changeIndex = (newIndex: number) => {setIndex(newIndex);}

    if (user.watch_history.length === 0)
        return (<div className="flex justify-center pt-5"><p>You haven&#39;t seen any films yet.</p></div>);
    return (<Pagination currenIndex={index} onClick={changeIndex} totalPage={3}>
        <MoviesCard movieSets={user.watch_history.map(m => movies[m.movie_id])}/>
    </Pagination>);
}

function CommentsTab({user}: {user: tUser}) {
    const allComments = comments.filter(comment => comment.author_id === user.id);

    return (<div className="max-w-3xl w-full mx-auto">
        <Comments user={user} comments={allComments}/>
    </div>);
}
