import {comments, tComment} from "@/types/comment";
import {tUser} from "@/types/user";
import React, {useState} from "react";

import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import "dayjs/locale/fr";
import SmallButton from "@/components/SmallButton";
import Pagination from "@/components/Pagination";
import {Button} from "@/components/Button";
import ProfilePicture from "@/components/ProfilePicture";
import {useAuth} from "@/context/AuthContext";
import {useModal} from "@/context/ModalContext";

dayjs.extend(relativeTime);
dayjs.locale("fr");

const MAX_COMMENT_SIZE = 300


export default function CommentSection() {
    const [actualComments, setComments] = useState(comments);
    const {user} = useAuth();
    const [index, setIndex] = useState(0);
    const {openModal} = useModal();

    const addNewComment = (newComment: tComment) => {setComments([...actualComments, newComment]);}
    const changeIndex = (newIndex: number) => {setIndex(newIndex);}

    return (<div className="mt-14 flex flex-col items-center mx-auto py-4 gap-4">
        <div className="border-b-5 border-b-yellow w-full mb-6">
            <h6 className="text-8xl">Comment</h6>
        </div>

        <Pagination currenIndex={index} totalPage={5} onClick={changeIndex}>
            <div className="flex flex-col-reverse gap-8 max-w-2xl">
                {actualComments.map((comment, index) => (<Comment key={index} currentUser={user} comment={comment}/>))}
                {user !== null ? <div className="flex gap-4 mb-2">
                    <ProfilePicture user={user}/>
                    <NewComment user={user} onSubmit={addNewComment}></NewComment>
                </div> : <button onClick={() => openModal({type: "signin"})} className="hover:underline font-extralight">Connectez-vous pour pouvoir poster un commentaire</button>}
            </div>
        </Pagination>
    </div>);
}

function Comment({comment, currentUser}: { comment: tComment, currentUser: tUser | null }) {
    let user: Partial<tUser>;

    if (currentUser && currentUser.id === comment.author_id)
        user = currentUser;
    else
        user = {id: comment.author_id, username: comment.author_username, firstname: comment.author_firstname, lastname: comment.author_lastname, profile_picture: comment.author_profile_pictures, color: comment.author_color};
    const [isCommentExpend, setIsExpendComment] = useState(false);

    return (<div className="w-full">
        <div className="flex gap-4">
            <ProfilePicture user={user}/>
            <div>
                <span className="text-bold">{user.username}</span>
                <p className="text-sm font-normal text-gray leading-4 mb-2">{dayjs.unix(comment.created_at).fromNow()}</p>
                <p className={isCommentExpend ? "" : "line-clamp-3"}>
                    {comment.comment}
                </p>
                {comment.comment.length > MAX_COMMENT_SIZE && (<SmallButton onClick={() => setIsExpendComment(!isCommentExpend)}>
                    {isCommentExpend ? "Reduire" : "Lire la suite"}</SmallButton>)}
            </div>
        </div>
    </div>);
}

function NewComment({user, onSubmit}: { user: tUser, onSubmit: (value: tComment) => void }) {
    const [expendComment, setExpendComment] = useState(false);
    const [comment, setComment] = useState("");
    const [canPost, setCanPost] = useState(false);

    const handleComment = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        setComment(e.target.value);
        setCanPost(e.target.value.length !== 0);
    }

    const handlePostComment = () => {
        setComment("");
        setCanPost(false);
        setExpendComment(false);

        const newComment: tComment = {
            id: Math.floor(Date.now() / 1000),
            author_id: user.id,
            author_username: user.username,
            author_firstname: user.firstname,
            author_lastname: user.lastname,
            author_profile_pictures: user.profile_picture,
            author_color: user.color,
            comment: comment,
            edited: false,
            created_at: Math.floor(Date.now() / 1000)
        }
        onSubmit(newComment);
    }

    return (<div className="flex flex-col items-center w-full gap-2">
        <textarea className="border w-full block px-3 py-1.5"
                  style={{resize: expendComment ? "vertical" : "none"}}
                  maxLength={1000} rows={expendComment ? 5 : 1}
                  placeholder={expendComment ? "" : "Écrire un commentaire..."}
                  onClick={() => setExpendComment(true)}
                  onChange={handleComment} value={comment}>
        </textarea>
        {expendComment &&
            <Button onClick={handlePostComment} disabled={!canPost} className="w-full">Publier le commentaire</Button>}
        {expendComment &&
            <SmallButton onClick={() => setExpendComment(false)}>Annuler</SmallButton>}
    </div>);
}
