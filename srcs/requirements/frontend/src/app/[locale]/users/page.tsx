"use client";

import React, {useEffect} from "react";
import {useRouter} from "@/i18n/navigation";
import useAuth from "@/contexts/AuthContext";
import MovieHistoryTab from "@/components/features/movie/MovieHistoryTab";
import CommentsProfileTab from "@/components/features/comment/CommentsProfileTab";
import UserProfile, {tTab} from "@/components/features/user/UserProfile";
import AuthProfileTab from "@/components/features/auth/AuthProfileTab";
import AvatarProfileTab from "@/components/features/user/AvatarProfileTab";
import ProfileTab from "@/components/features/user/UserProfileTab";
import APITab from "@/components/features/auth/APITab";

export default function Page() {
    const {user, loading, updateUser} = useAuth();
    const router = useRouter();
    const tabs: tTab = [{name: "history", comp: MovieHistoryTab}, {name: "comments", comp: CommentsProfileTab}, {name: "api", comp: APITab}];

    useEffect(() => {
        if (!loading && !user)
            router.push("/");
    }, [user, loading, router]);


    if (!user)
        return null;

    if (user.oauth_method)
        tabs.unshift({name: "avatar", comp: AvatarProfileTab});
    else {
        tabs.unshift({name: "auth", comp: AuthProfileTab});
        tabs.unshift({name: "profile", comp: ProfileTab});
    }

    return <UserProfile user={user} updateUserAction={updateUser} tabs={tabs} />;
}
