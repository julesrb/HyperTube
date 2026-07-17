"use client";

import React, {useEffect, useState} from "react";
import UserProfile, {tTab} from "@/components/features/user/UserProfile";
import MovieHistoryTab from "@/components/features/movie/MovieHistoryTab";
import useHandleError from "@/hooks/useHandleError";
import {useUser} from "@/services/users.service";
import CommentsProfileTab from "@/components/features/comment/CommentsProfileTab";
import {ApiError} from "@/services/ApiError";
import {useParams} from "next/navigation";

export default function Page() {
    const params = useParams();
    const userId = params.id as string;
    const [errorNode, setErrorNode] = useState<React.ReactNode>(null);
    const tabs: tTab = [{name: "history", comp: MovieHistoryTab}, {name: "comments", comp: CommentsProfileTab}];
    const handleError = useHandleError();
    const {data, error} = useUser(userId);

    useEffect(() => {
        if (error) {
            const node = handleError(error as ApiError, "User");
            // eslint-disable-next-line react-hooks/set-state-in-effect
            setErrorNode(node);
        }
    }, [error, handleError]);

    if (errorNode || !data)
        return (errorNode);

    return <UserProfile user={data.data} tabs={tabs} />;
}
