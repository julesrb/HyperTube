import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import {iUser} from "@/types/user";
import React, {useEffect, useRef, useState} from "react";
import useModal from "@/contexts/ModalContext";
import {useTranslations} from "next-intl";
import {iComment, iCommentDetails} from "@/types/comment";
import MovieCard from "@/components/features/movie/MovieCard";
import {Link} from "@/i18n/navigation";
import ProfilePicture from "@/components/features/user/ProfilePicture";
import {EditIcon, TrashIcon} from "@/components/Icons";
import TextButton from "@/components/ui/Button/TextButton";
import SecondaryButton from "@/components/ui/Button/SecondaryButton";
import Button from "@/components/ui/Button/Button";
import IconButton from "@/components/ui/Button/IconButton";

dayjs.extend(relativeTime);

export default function Comment({comment, currentUser, updateComment, deleteComment, previousCommentMovieId}: {comment: iComment | iCommentDetails, currentUser?: iUser, updateComment: (comment: iComment, newContent: string) => void, deleteComment: (commentId: number, movieId?: string) => Promise<void>, previousCommentMovieId?: string}) {
    let user: iUser;
    const [showSettingBtn, setShowSettingBtn] = useState(false);
    const [editMode, setEditMode] = useState(false);
    const {openModal} = useModal();
    const t = useTranslations("comments");

    if (currentUser && currentUser.id === comment.user.id)
        user = currentUser;
    else
        user = comment.user;

    return (<div className="w-full"
            onMouseEnter={() => setShowSettingBtn(true)}
            onMouseLeave={() => setShowSettingBtn(false)}>
        {"movie" in comment && previousCommentMovieId != comment.movie.imdb_id && <div className="flex justify-center mb-3">
            <MovieCard user={currentUser} className="aspect-21/9" movie={comment.movie}/></div>}
        <div className={"flex gap-2 sm:gap-4" + ((!updateComment) ? " flex-col sm:flex-row mx-4" : "")}>
            <Link href={`/users/${user.id}`}><ProfilePicture user={user} /></Link>
            <div className="w-full">
                <div className="flex justify-between w-full">
                    <div>
                        <Link href={`/users/${user.id}`} className="text-bold hover:underline">{user.username}</Link>
                        <p className="text-sm font-normal text-gray leading-4 mb-2">{dayjs(comment.updated_at).fromNow()} {comment.edited && ` • ${t("edited")}`}</p>
                    </div>
                    {
                        (currentUser && comment.user.id === currentUser.id && showSettingBtn) &&
                        <div className="flex gap-1">
                            <IconButton onClick={() => setEditMode(true)} className="uppercase font-condensed text-2xl">
                                {(color: string) => <EditIcon color={color}/>}
                            </IconButton>
                            <IconButton onClick={() => {
                                setEditMode(false);
                                openModal({type: "delete-confirmation", deleteObjId: comment.id, deleteFunc: () => deleteComment(comment.id, "movie" in comment ? comment.movie.imdb_id : undefined)});
                            }} className="uppercase font-condensed text-2xl" hoverColor="red">
                                {(color: string) => <TrashIcon color={color}/>}
                            </IconButton>
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
        {isClamped && (<TextButton onClick={() => setIsExpendComment(!isCommentExpend)}>
            {isCommentExpend ? t("collapse") : t("readMore")}</TextButton>)}
    </div>);
}

function CommentTextEdit({comment, setEditMode, updateComment}: {comment: iComment | iCommentDetails, setEditMode: (newEditMode: boolean) => void, updateComment: (comment: iComment, newContent: string) => void}) {
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
        updateComment(comment, newEditedComment.trim());
        setEditMode(false);
    };

    return (<div className="flex flex-col gap-3">
        <textarea ref={textareaRef} value={newEditedComment}
                  onInput={autoResize}
                  className="w-full resize-none font-light"
                  onKeyDown={(e) => {
                      if (e.key === "Enter" && !e.shiftKey) {
                          e.preventDefault();
                          if (newEditedComment.trim().length > 0 && newEditedComment.trim() !== comment.content)
                              saveChange();
                      }
                  }}
                  onChange={(e) => setNewEditedComment(e.target.value)} />
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
