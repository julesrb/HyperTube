import React, {useRef, useState} from "react";
import {useTranslations} from "next-intl";
import useApiMutation from "@/hooks/useApiMutation";
import {addCommentCache, postComment} from "@/services/comments.service";
import {iCommentDetails} from "@/types/comment";
import Button from "@/components/ui/Button/Button";
import TextButton from "@/components/ui/Button/TextButton";
import {iUser} from "@/types/user";
import {iMovie} from "@/types/movie";
import {useQueryClient} from "@tanstack/react-query";

export default function NewComment({user, movie}: {user: iUser, movie: iMovie}) {
    const [expendComment, setExpendComment] = useState(false);
    const [comment, setComment] = useState("");
    const t = useTranslations("comments");
    const [errors, setErrors] = useState<Record<string, string>>({});
    const {execute} = useApiMutation(setErrors);
    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const queryClient = useQueryClient();

    const reset = () => {
        setComment("");
        setExpendComment(false);
        textareaRef?.current?.blur();
    };

    const handleComment = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        if (expendComment)
            setComment(e.target.value);
    }

    const handlePostComment = () => {
        const makePostRequest = async () => {
            return await execute((locale) => postComment(locale, movie.imdb_id, comment.trim()));
        };

        makePostRequest().then((data) => {
            if (data) {
                data.data.user = user;
                const newComment = data.data as iCommentDetails;
                addCommentCache(queryClient, newComment, movie, newComment.user.id);
            }
            reset();
        })
    }

    return (<div className="flex flex-col items-center w-full gap-2">
        <textarea ref={textareaRef} className={"border w-full block px-3 py-1.5" + (errors["content"] ? " border-red text-red" : "")}
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
        {errors["content"] && <span className="text-red text-xs">{errors["content"]}</span>}
        {expendComment && <Button onClick={handlePostComment} disabled={comment.trim().length <= 0} className="w-full">{t("publishComment")}</Button>}
        {expendComment && <TextButton onClick={reset}>{t("cancel")}</TextButton>}
    </div>);
}
