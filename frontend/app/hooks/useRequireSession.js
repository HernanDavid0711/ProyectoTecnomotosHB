"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getSession } from "../lib/api";

export default function useRequireSession(expectedRole) {
  const router = useRouter();
  const [session, setSession] = useState(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const current = getSession();
    if (!current) {
      router.replace("/login");
      return;
    }
    if (expectedRole && current.user?.rol !== expectedRole) {
      router.replace(current.user?.rol === "creador_ordenes" ? "/creador/ordenes" : "/admin/dashboard");
      return;
    }
    setSession(current);
    setReady(true);
  }, [expectedRole, router]);

  return { session, ready };
}
