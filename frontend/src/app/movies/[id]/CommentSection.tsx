import {comments, tComment} from "@/types/comment";
import {tUser} from "@/types/user";
import React, {useEffect, useRef, useState} from "react";

import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import "dayjs/locale/fr";
import Pagination from "@/components/Pagination";
import {Button, SecondaryButton, SmallButton} from "@/components/Buttons";
import ProfilePicture from "@/components/ProfilePicture";
import {useAuth} from "@/context/AuthContext";
import {useModal} from "@/context/ModalContext";
import {EditIcon, TrashIcon} from "@/components/Icons";
import {movies, tMovie} from "@/types/movie";
import {MovieCard} from "@/components/MovieCard";
import {useNotification} from "@/context/NotificationContext";
import {successMessages} from "@/types/message";

dayjs.extend(relativeTime);
dayjs.locale("fr");


export function CommentSection({movie}: {movie: tMovie}) {
    const {user} = useAuth();
    const {addNotification} = useNotification();
    const {openModal} = useModal();
    const [actualComments, setComments] = useState(comments);
    const addNewComment = (newComment: tComment) => {setComments([...actualComments, newComment]);}
    const updateComment = (commentId: number, newContent: string) => {
        setComments(actualComments.map((comment) => {
            if (comment.id === commentId) {
                const newComment = structuredClone(comment);
                newComment.comment = newContent.replace('\n\n', '\n');
                newComment.edited = true;
                return newComment;
            }
            else
                return comment;
        }));
        addNotification(successMessages.commentChange, "success");
    }
    const deleteComment = (commentId: number) => {setComments(actualComments.filter(c => c.id !== commentId));}

    return (<div className="mt-14 flex flex-col items-center py-4 gap-4">
        <div className="border-b-5 border-b-yellow w-full mb-6">
            <h6 className="text-8xl">Comment</h6>
        </div>
        <div className="max-w-2xl w-full">
            {user !== null ? <div className="flex gap-4 mb-8 w-full">
                <ProfilePicture user={user}/>
                <NewComment user={user} onSubmit={addNewComment} movie={movie}></NewComment>
            </div> : <button onClick={() => openModal({type: "signin"})} className="hover:underline font-extralight">Connectez-vous pour pouvoir poster un commentaire</button>}
            <Comments user={user} comments={comments} updateComment={updateComment} deleteComment={deleteComment}/>
        </div>
    </div>);
}

export function Comments({user, comments, updateComment, deleteComment}: {user: tUser | null, comments: tComment[], updateComment?: (commentId: number, newContent: string) => void, deleteComment?: (commentId: number) => void}) {
    const [index, setIndex] = useState(0);
    const changeIndex = (newIndex: number) => {setIndex(newIndex);}

    return (<Pagination currenIndex={index} totalPage={5} onClick={changeIndex}>
        <div className="flex flex-col-reverse gap-8">
            {comments.map((comment, index) => (<Comment key={index} currentUser={user} comment={comment} updateComment={updateComment} deleteComment={deleteComment}/>))}
        </div>
    </Pagination>);
}

function Comment({comment, currentUser, updateComment, deleteComment}: { comment: tComment, currentUser: tUser | null, updateComment?: (commentId: number, newContent: string) => void, deleteComment?: (commentId: number) => void}) {
    let user: Partial<tUser>;
    const [showSettingBtn, setShowSettingBtn] = useState(false);
    const [editMode, setEditMode] = useState(false);
    const [hoverTrash, setHoverTrash] = useState(false);
    const {openModal} = useModal();
    const movie = movies.find(m => m.id === comment.movie_id);

    if (currentUser && currentUser.id === comment.author_id)
        user = currentUser;
    else
        user = {id: comment.author_id, username: comment.author_username, firstname: comment.author_firstname, lastname: comment.author_lastname, profile_picture: comment.author_profile_pictures, color: comment.author_color};

    return (<div className="w-full"
            onMouseEnter={() => setShowSettingBtn(true)}
            onMouseLeave={() => setShowSettingBtn(false)}>
        <div className="flex gap-4">
            {(!updateComment && movie) && <MovieCard user={currentUser} className="max-w-50" showTitle={false} movie={movie} />}
            <ProfilePicture user={user}/>
            <div className="w-full">
                <div className="flex justify-between w-full">
                    <div>
                        <span className="text-bold">{user.username}</span>
                        <p className="text-sm font-normal text-gray leading-4 mb-2">{dayjs.unix(comment.created_at).fromNow()} {comment.edited && " • Édité"}</p>
                    </div>
                    {/* todo mby replace icon by text 'edit', 'remove' */}
                    {
                        (updateComment && currentUser !== null && comment.author_id === currentUser.id && showSettingBtn) &&
                        <div className="flex gap-1">
                            <button
                                className="uppercase font-condensed text-2xl"
                                onClick={() => setEditMode(true)}><EditIcon /></button>
                            <button
                                className="uppercase font-condensed text-2xl"
                                onMouseLeave={() => setHoverTrash(false)}
                                onMouseEnter={() => setHoverTrash(true)}
                                onClick={() => {
                                    setEditMode(false);
                                    openModal({type: "delete-comment", commentId: comment.id, deleteComment: deleteComment});
                                }}><TrashIcon color={hoverTrash ? "red" : "black"}/></button>
                        </div>
                    }
                </div>
                {editMode && updateComment ?
                    <CommentTextEdit comment={comment} setEditMode={setEditMode} updateComment={updateComment}/>
                    : <CommentText comment={comment}/>
                }
            </div>
        </div>
    </div>);
}

function CommentText({comment}: {comment: tComment}) {
    const [isCommentExpend, setIsExpendComment] = useState(false);
    const [isClamped, setIsClamped] = useState(false);
    const textRef = useRef<HTMLParagraphElement>(null);

    useEffect(() => {
        const el = textRef.current;
        if (!el) return;
        const checkClamp = () => {
            setIsClamped(el.scrollHeight > el.clientHeight);
        };
        checkClamp();
        window.addEventListener("resize", checkClamp);

        return () => window.removeEventListener("resize", checkClamp);
    }, [comment]);

    return (<div>
        <p ref={textRef} className={"whitespace-pre-line " + (isCommentExpend ? "" : "line-clamp-3")}>
            {comment.comment}
        </p>
        {isClamped && (<SmallButton onClick={() => setIsExpendComment(!isCommentExpend)}>
            {isCommentExpend ? "Reduire" : "Lire la suite"}</SmallButton>)}
    </div>);
}

function CommentTextEdit({comment, setEditMode, updateComment}: {comment: tComment, setEditMode: (newEditMode: boolean) => void, updateComment: (commentId: number, newContent: string) => void}) {
    const [newEditedComment, setNewEditedComment] = useState(comment.comment);
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    useEffect(() => {
        const el = textareaRef.current;
        if (!el) return;
        el.style.height = "auto";
        el.style.height = el.scrollHeight + "px";
        el.focus();
        el.setSelectionRange(el.value.length, el.value.length);
    }, [comment]);

    const autoResize = () => {
        const el = textareaRef.current;
        if (!el) return;
        el.style.height = "auto";
        el.style.height = el.scrollHeight + "px";
    };

    const saveChange = () => {
        updateComment(comment.id, newEditedComment.trim());
        setEditMode(false);
    };

    return (<div className="flex flex-col gap-3">
        <textarea ref={textareaRef} value={newEditedComment}
                  onInput={autoResize}
                  className="w-full resize-none font-sans"
                  onKeyDown={(e) => {
                      if (e.key === "Enter" && !e.shiftKey) {
                          e.preventDefault();
                          saveChange();
                      }
                  }}
                  onChange={(e) => setNewEditedComment(e.target.value)}></textarea>
        <div className="flex gap-2">
            <Button
                disabled={newEditedComment.trim().length <= 0 || newEditedComment.trim() === comment.comment}
                onClick={saveChange}>save change</Button>
            <SecondaryButton className="w-40" onClick={() => {
                setEditMode(false);
                setNewEditedComment(comment.comment);
            }}>cancel</SecondaryButton>
        </div>
    </div>);
}

function NewComment({user, movie, onSubmit}: { user: tUser, movie: tMovie, onSubmit: (value: tComment) => void }) {
    const [expendComment, setExpendComment] = useState(false);
    const [comment, setComment] = useState("");

    const handleComment = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        if (expendComment)
            setComment(e.target.value);
    }

    const handlePostComment = () => {
        const newComment: tComment = {
            id: Math.floor(Date.now() / 1000),
            movie_id: movie.id,
            author_id: user.id,
            author_username: user.username,
            author_firstname: user.firstname,
            author_lastname: user.lastname,
            author_profile_pictures: user.profile_picture,
            author_color: user.color,
            comment: comment.trim(),
            edited: false,
            created_at: Math.floor(Date.now() / 1000)
        }
        setComment("");
        setExpendComment(false);
        onSubmit(newComment);
    }

    return (<div className="flex flex-col items-center w-full gap-2">
        <textarea className="border w-full block px-3 py-1.5"
                  style={{resize: expendComment ? "vertical" : "none"}}
                  maxLength={1000} rows={expendComment ? 5 : 1}
                  placeholder={expendComment ? "" : "Écrire un commentaire..."}
                  onClick={() => setExpendComment(true)}
                  onKeyDown={(e) => {
                      if (comment.trim().length > 0 && e.key === "Enter" && !e.shiftKey) {
                          e.preventDefault();
                          handlePostComment();
                      }
                  }}
                  onChange={handleComment} value={comment}>
        </textarea>
        {expendComment &&
            <Button onClick={handlePostComment} disabled={comment.trim().length <= 0} className="w-full">Publier le commentaire</Button>}
        {expendComment &&
            <SmallButton onClick={() => setExpendComment(false)}>Annuler</SmallButton>}
    </div>);
}
