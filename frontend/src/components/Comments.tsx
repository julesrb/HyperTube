import {comments, tComment} from "@/types/comment";
import {tUser} from "@/types/user";
import React, {useEffect, useRef, useState} from "react";

import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import "dayjs/locale/fr";
import "dayjs/locale/en";
import "dayjs/locale/de";
import Pagination from "@/components/Pagination";
import {Button, SecondaryButton, SmallButton} from "@/components/Buttons";
import ProfilePicture from "@/components/ProfilePicture";
import {useAuth} from "@/context/AuthContext";
import {useModal} from "@/context/ModalContext";
import {EditIcon, TrashIcon} from "@/components/Icons";
import {movies, tMovie} from "@/types/movie";
import {MovieCard} from "@/components/MovieCard";
import {useNotification} from "@/context/NotificationContext";
import {useLocale, useTranslations} from "next-intl";

dayjs.extend(relativeTime);


export function CommentSection({movie}: {movie: tMovie}) {
    const {user} = useAuth();
    const {addNotification} = useNotification();
    const {openModal} = useModal();
    const [actualComments, setComments] = useState(comments);
    const t = useTranslations("comments");
    const tSuccess = useTranslations("notifications.success");
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
        addNotification(tSuccess("commentChange"), "success");
    }
    const deleteComment = (commentId: number) => {setComments(actualComments.filter(c => c.id !== commentId));}

    return (<div className="mx-auto max-w-2xl w-9/10 sm:w-full flex flex-col items-center gap-7">
        <div className="w-full">
            <h1 className="text-center">{t("title")}</h1>
            <div className="flex h-2 sm:h-4 mt-1 sm:mt-2 w-full">
                <div className="size-full bg-yellow"></div>
                <div className="size-full bg-pink"></div>
                <div className="size-full bg-green"></div>
                <div className="size-full bg-purple"></div>
                <div className="size-full bg-blue"></div>
                <div className="size-full bg-red"></div>
            </div>
        </div>
        <div className="w-full text-center">
            {
                user !== null ?
                <div className="flex gap-2 sm:gap-4">
                    <ProfilePicture user={user}/>
                    <NewComment user={user} onSubmit={addNewComment} movie={movie}></NewComment>
                </div> :
                <SmallButton onClick={() => openModal({type: "signin"})}>{t("signInToComment")}</SmallButton>
            }
        </div>
        <Comments user={user} comments={actualComments} updateComment={updateComment} deleteComment={deleteComment}/>
    </div>);
}

export function Comments({user, comments, updateComment, deleteComment}: {user: tUser | null, comments: tComment[], updateComment?: (commentId: number, newContent: string) => void, deleteComment?: (commentId: number) => void}) {
    const [index, setIndex] = useState(0);
    const locale = useLocale();
    const t = useTranslations("comments");
    if (locale === "fr")
        dayjs.locale("fr");
    else if (locale === "en")
        dayjs.locale("en");
    else
        dayjs.locale("de");
    const changeIndex = (newIndex: number) => {setIndex(newIndex);}

    if (comments.length === 0)
        return (<p className="small-text">{t(deleteComment === undefined ? "noCommentsYet" : "noCommentsPrompt")}</p>);

    return (<Pagination currenIndex={index} totalPage={5} onClick={changeIndex}>
        <div className="flex flex-col-reverse gap-6">
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
    const t = useTranslations("comments");

    if (currentUser && currentUser.id === comment.author_id)
        user = currentUser;
    else
        user = {id: comment.author_id, username: comment.author_username, firstname: comment.author_firstname, lastname: comment.author_lastname, profile_picture: comment.author_profile_pictures, color: comment.author_color};

    return (<div className="w-full"
            onMouseEnter={() => setShowSettingBtn(true)}
            onMouseLeave={() => setShowSettingBtn(false)}>
        {(!updateComment && movie) && <div className="flex justify-center mb-3">
            <MovieCard user={currentUser} className="aspect-21/9" showTitle={false} movie={movie} /></div>}
        <div className={"flex gap-2 sm:gap-4" + ((!updateComment && movie) ? " flex-col sm:flex-row mx-4" : "")}>
            <ProfilePicture user={user}/>
            <div className="w-full">
                <div className="flex justify-between w-full">
                    <div>
                        <span className="text-bold">{user.username}</span>
                        <p className="text-sm font-normal text-gray leading-4 mb-2">{dayjs.unix(comment.created_at).fromNow()} {comment.edited && ` • ${t("edited")}`}</p>
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
                <div className="leading-tight sm:leading-normal">
                    {editMode && updateComment ?
                        <CommentTextEdit comment={comment} setEditMode={setEditMode} updateComment={updateComment}/>
                        : <CommentText comment={comment}/>
                    }
                </div>
            </div>
        </div>
    </div>);
}

function CommentText({comment}: {comment: tComment}) {
    const [isCommentExpend, setIsExpendComment] = useState(false);
    const [isClamped, setIsClamped] = useState(false);
    const textRef = useRef<HTMLParagraphElement>(null);
    const t = useTranslations("comments");

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
            {isCommentExpend ? t("collapse") : t("readMore")}</SmallButton>)}
    </div>);
}

function CommentTextEdit({comment, setEditMode, updateComment}: {comment: tComment, setEditMode: (newEditMode: boolean) => void, updateComment: (commentId: number, newContent: string) => void}) {
    const [newEditedComment, setNewEditedComment] = useState(comment.comment);
    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const t = useTranslations("comments");

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
                      if (e.key === "Enter" && !e.shiftKey && newEditedComment.trim().length > 0 && newEditedComment.trim() !== comment.comment) {
                          e.preventDefault();
                          saveChange();
                      }
                  }}
                  onChange={(e) => setNewEditedComment(e.target.value)}></textarea>
        <div className="flex gap-2">
            <Button className="xl:px-6"
                disabled={newEditedComment.trim().length <= 0 || newEditedComment.trim() === comment.comment}
                onClick={saveChange}>{t("saveChange")}</Button>
            <SecondaryButton className="w-30 xl:w-40" onClick={() => {
                setEditMode(false);
                setNewEditedComment(comment.comment);
            }}>{t("cancel")}</SecondaryButton>
        </div>
    </div>);
}

function NewComment({user, movie, onSubmit}: { user: tUser, movie: tMovie, onSubmit: (value: tComment) => void }) {
    const [expendComment, setExpendComment] = useState(false);
    const [comment, setComment] = useState("");
    const t = useTranslations("comments");

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
                  placeholder={expendComment ? "" : t("writeComment")}
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
            <Button onClick={handlePostComment} disabled={comment.trim().length <= 0} className="w-full">{t("publishComment")}</Button>}
        {expendComment &&
            <SmallButton onClick={() => setExpendComment(false)}>{t("cancel")}</SmallButton>}
    </div>);
}
