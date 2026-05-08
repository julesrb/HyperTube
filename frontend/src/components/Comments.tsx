import {iComment} from "@/types/comment";
import {iUser} from "@/types/user";
import React, {useEffect, useRef, useState} from "react";

import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import "dayjs/locale/fr";
import "dayjs/locale/en";
import "dayjs/locale/de";
import Pagination, {computeTotalPage} from "@/components/Pagination";
import {Button, SecondaryButton, SmallButton} from "@/components/Buttons";
import ProfilePicture from "@/components/ProfilePicture";
import {useAuth} from "@/context/AuthContext";
import {useModal} from "@/context/ModalContext";
import {EditIcon, TrashIcon} from "@/components/Icons";
import {iMovie} from "@/types/movie";
import {MovieCard} from "@/components/MovieCard";
import {useNotification} from "@/context/NotificationContext";
import {useLocale, useTranslations} from "next-intl";
import {getComments} from "@/services/comments";
import {getMovie} from "@/services/movies";

dayjs.extend(relativeTime);


export function CommentSection({movie}: {movie: iMovie}) {
    const {user} = useAuth();
    const {addNotification} = useNotification();
    const {openModal} = useModal();
    const [actualComments, setComments] = useState<iComment[]>([]);
    const [index, setIndex] = useState(0);
    const [totalPage, setTotalPage] = useState(1);
    useEffect(() => {
        async function loadComments() {
            try {
                const data = await getComments(movie.imdb_id);
                setComments(data.data);
                computeTotalPage(data, setTotalPage);
            } catch (error) {
                console.error(error);
            }
        }
        loadComments().then(r => console.log(r));
    }, [movie.imdb_id]);
    const t = useTranslations("comments");
    const tSuccess = useTranslations("notifications.success");
    const addNewComment = (newComment: iComment) => {setComments([...actualComments, newComment]);}
    const updateComment = (commentId: number, newContent: string) => {
        setComments(actualComments.map((comment) => {
            if (comment.id === commentId) {
                const newComment = structuredClone(comment);
                newComment.content = newContent.replace('\n\n', '\n');
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
        <Comments user={user} totalPage={totalPage} comments={actualComments} updateComment={updateComment} deleteComment={deleteComment} index={index} setIndex={setIndex}/>
    </div>);
}

export function Comments({user, totalPage, index, setIndex, comments, updateComment, deleteComment}: {user: iUser | null, totalPage: number, index: number, setIndex: (idx: number) => void, comments: iComment[], updateComment?: (commentId: number, newContent: string) => void, deleteComment?: (commentId: number) => void}) {
    const locale = useLocale();
    const t = useTranslations("comments");
    if (locale === "fr")
        dayjs.locale("fr");
    else if (locale === "en")
        dayjs.locale("en");
    else
        dayjs.locale("de");
    const changeIndex = (newIndex: number) => {setIndex(newIndex);}

    if (!comments || comments.length === 0)
        return (<p className="small-text mb-10">{t(deleteComment === undefined ? "noCommentsYet" : "noCommentsPrompt")}</p>);

    return (<Pagination currenIndex={index} totalPage={totalPage} onClick={changeIndex}>
        <div className="flex flex-col-reverse gap-6">
            {comments.map((comment, index) => (<Comment key={index} currentUser={user} comment={comment} updateComment={updateComment} deleteComment={deleteComment}/>))}
        </div>
    </Pagination>);
}

function Comment({comment, currentUser, updateComment, deleteComment}: { comment: iComment, currentUser: iUser | null, updateComment?: (commentId: number, newContent: string) => void, deleteComment?: (commentId: number) => void}) {
    let user: iUser;
    const [showSettingBtn, setShowSettingBtn] = useState(false);
    const [editMode, setEditMode] = useState(false);
    const [hoverTrash, setHoverTrash] = useState(false);
    const [movie, setMovie] = useState<null | iMovie>(null);
    const {openModal} = useModal();
    const t = useTranslations("comments");

    useEffect(() => {
        async function loadMovie() {
            try {
                const data = await getMovie(comment.movie_id);
                setMovie(data.data);
            } catch (error) {
                console.error(error);
            }
        }
        loadMovie().then(r => console.log(r));
    }, [comment.movie_id]);

    if (currentUser && currentUser.id === comment.user.id)
        user = currentUser;
    else
        user = comment.user;

    return (<div className="w-full"
            onMouseEnter={() => setShowSettingBtn(true)}
            onMouseLeave={() => setShowSettingBtn(false)}>
        {(!updateComment && movie) && <div className="flex justify-center mb-3">
            <MovieCard user={currentUser} className="aspect-21/9" showTitle={false} movie={movie} /></div>}
        <div className={"flex gap-2 sm:gap-4" + ((!updateComment) ? " flex-col sm:flex-row mx-4" : "")}>
            <ProfilePicture user={user}/>
            <div className="w-full">
                <div className="flex justify-between w-full">
                    <div>
                        <span className="text-bold">{user.username}</span>
                        <p className="text-sm font-normal text-gray leading-4 mb-2">{dayjs.unix(comment.updated_at).fromNow()} {comment.edited && ` • ${t("edited")}`}</p>
                    </div>
                    {/* todo mby replace icon by text 'edit', 'remove' */}
                    {
                        (updateComment && currentUser !== null && comment.user.id === currentUser.id && showSettingBtn) &&
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

function CommentText({comment}: {comment: iComment}) {
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
            {comment.content}
        </p>
        {isClamped && (<SmallButton onClick={() => setIsExpendComment(!isCommentExpend)}>
            {isCommentExpend ? t("collapse") : t("readMore")}</SmallButton>)}
    </div>);
}

function CommentTextEdit({comment, setEditMode, updateComment}: {comment: iComment, setEditMode: (newEditMode: boolean) => void, updateComment: (commentId: number, newContent: string) => void}) {
    const [newEditedComment, setNewEditedComment] = useState(comment.content);
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
                      if (e.key === "Enter" && !e.shiftKey && newEditedComment.trim().length > 0 && newEditedComment.trim() !== comment.content) {
                          e.preventDefault();
                          saveChange();
                      }
                  }}
                  onChange={(e) => setNewEditedComment(e.target.value)}></textarea>
        <div className="flex gap-2">
            <Button className="xl:px-6"
                disabled={newEditedComment.trim().length <= 0 || newEditedComment.trim() === comment.content}
                onClick={saveChange}>{t("saveChange")}</Button>
            <SecondaryButton className="w-30 xl:w-40" onClick={() => {
                setEditMode(false);
                setNewEditedComment(comment.content);
            }}>{t("cancel")}</SecondaryButton>
        </div>
    </div>);
}

function NewComment({user, movie, onSubmit}: { user: iUser, movie: iMovie, onSubmit: (value: iComment) => void }) {
    const [expendComment, setExpendComment] = useState(false);
    const [comment, setComment] = useState("");
    const t = useTranslations("comments");

    const handleComment = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        if (expendComment)
            setComment(e.target.value);
    }

    const handlePostComment = () => {
        const newComment: iComment = {
            id: Math.floor(Date.now() / 1000),
            movie_id: movie.imdb_id,
            user: user,
            content: comment.trim(),
            edited: false,
            updated_at: Math.floor(Date.now() / 1000)
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
