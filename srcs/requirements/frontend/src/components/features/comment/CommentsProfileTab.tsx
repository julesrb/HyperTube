import {useEffect, useState} from "react";
import {iUser} from "@/types/user";
import {iComment, iCommentDetails} from "@/types/comment";
import {useProfileComments} from "@/services/comments.service";
import computeTotalPage from "@/utils/computeTotalPage";
import Comments from "@/components/features/comment/Comments";
import useAuth from "@/contexts/AuthContext";
import {useUserFilmHistory} from "@/services/users.service";
import addProgressToMovie from "@/utils/addProgressToMovie";

export default function CommentsProfileTab({user}: {user: iUser}) {
    const [index, setIndex] = useState(0);
    const {data} = useProfileComments(user.id, index);
    const {user: currentUser} = useAuth();
    const [actualComments, setComments] = useState<iComment[]>([]);
    const [totalPage, setTotalPage] = useState(1);
    const {data: watchHistory} = useUserFilmHistory(user.id);

    useEffect(() => {
        if (!data)
            return;
        const dataWithUser: iCommentDetails[] = [];
        data.data.map(comment => dataWithUser.push({...comment, user: user}))
        addProgressToMovie(watchHistory?.data, dataWithUser)
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setComments(dataWithUser);
        setTotalPage(computeTotalPage(data));
    }, [data, user, watchHistory]);

    return (<div className="max-w-3xl w-full mx-auto">
        <Comments currentUser={currentUser} comments={actualComments} index={index} setIndex={setIndex} totalPage={totalPage} profilePage={true}/>
    </div>);
}
