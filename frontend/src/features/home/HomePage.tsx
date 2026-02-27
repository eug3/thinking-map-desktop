"use client";
import { useState } from "react";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Command, Github, Loader } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { toast } from "sonner";
import { useNavigate } from "react-router-dom";
import { createMap } from "@/api/map";
import type { CreateMapRequest } from "@/types/map";
import SettingsDialog from "@/components/SettingsDialog";

const PROBLEM_TYPES = [
  { value: '研究型', label: '研究型', description: '深入探索和分析特定主题' },
  { value: '创意型', label: '创意型', description: '发散思维，寻找创新解决方案' },
  { value: '分析型', label: '分析型', description: '系统分析数据和现象' },
  { value: '规划型', label: '规划型', description: '制定策略和执行计划' }
];

const PROBLEM_EXAMPLES = {
  研究型: '评估人工智能在教育领域的应用效果？',
  创意型: '设计一个创新的城市共享单车系统',
  分析型: '分析全球供应链中断对电子产品市场的影响',
  规划型: 'python入门学习规划'
};
const GITHUB_URL = "https://github.com/PGshen/thinking-map";

export default function HomePage() {
  const [problem, setProblem] = useState('');
  const [problemType, setProblemType] = useState('研究型');

  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [title, setTitle] = useState('');
  const [goal, setGoal] = useState('');
  const [isLoading, _setIsLoading] = useState(false);

  const navigate = useNavigate();

  const handleCreateMap = async () => {
    if (!problem.trim()) {
      toast.error("问题描述不能为空");
      return;
    }
    const params: CreateMapRequest = {
      title: title.trim() || problem.trim().substring(0, 50),
      problem: problem.trim(),
      problemType: problemType || undefined,
      target: goal.trim() || undefined,
      keyPoints: undefined,
      constraints: undefined,
    };
    try {
      const res = await createMap(params);
      if (res && res.id) {
        setIsDialogOpen(false);
        navigate(`/workspace/${res.id}`);
      } else {
        toast.error("创建思维导图失败");
      }
    } catch (e) {
      toast.error("创建思维导图失败");
    }
  };

  const handleSubmit = () => {
    if (problem.length < 6) {
      toast.info("问题描述至少需要6个字，请补充更多细节。");
      return;
    }
    setIsDialogOpen(true);
  };

  return (
    <div className="container mx-auto px-4 py-8 flex flex-col items-center justify-center min-h-screen">
      <SettingsDialog />
      <a
        href={GITHUB_URL}
        target="_blank"
        rel="noopener noreferrer"
        aria-label="打开 GitHub 仓库"
        className="fixed top-3 right-3 inline-flex items-center gap-1 rounded-full bg-black text-white px-2 py-2 shadow hover:bg-gray-900"
      >
        <Github className="h-4 w-4" />
        <span className="text-xs font-medium">GitHub</span>
      </a>

      {/* 品牌展示区 */}
      <div className="text-center mb-8">
        <div className="flex mb-2 items-end justify-center">
          <div className="mr-4 flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-lg">
            <Command className="size-6" />
          </div>
          <h1 className="text-4xl font-bold">ThinkingMap Desktop</h1>
        </div>
        <p className="text-lg text-gray-600">可视化解构您的思维历程，让问题解决变得透明和可控</p>
      </div>

      {/* 问题输入区 */}
      <div className="w-full max-w-2xl space-y-3">
        <div className="relative">
          <Textarea
            placeholder="请输入您的问题（50-200字最佳）..."
            value={problem}
            onChange={(e) => setProblem(e.target.value)}
            className="min-h-[120px] pr-24"
          />
          <div className="absolute bottom-2 right-2 flex items-center space-x-2">
            <span className="text-sm text-gray-500">
              {problem.length} / 200 字
              {problem.length > 200 && (
                <span className="text-yellow-500 ml-2">字数超出建议范围</span>
              )}
            </span>
            <Button
              onClick={handleSubmit}
              disabled={!problem.trim() || !problemType}
              className="ml-2 cursor-pointer"
              size="sm"
            >
              开始分析
            </Button>
          </div>
        </div>

        <div className="space-y-3">
          <Label className="text-base">选择问题类型</Label>
          <RadioGroup
            value={problemType}
            onValueChange={setProblemType}
            className="grid grid-cols-1 md:grid-cols-2 gap-1"
          >
            {PROBLEM_TYPES.map((type) => (
              <div key={type.value} className="flex items-center space-x-2 p-1 rounded-lg hover:bg-slate-50">
                <RadioGroupItem value={type.value} id={type.value} />
                <Label htmlFor={type.value} className="flex-1 cursor-pointer">
                  <span className="font-medium">{type.label}</span>
                  <span className="text-gray-500 ml-2">- {type.description}</span>
                </Label>
              </div>
            ))}
          </RadioGroup>
        </div>

        {problemType && (
          <div className="text-sm text-gray-600">
            <p>示例问题：{PROBLEM_EXAMPLES[problemType as keyof typeof PROBLEM_EXAMPLES]}</p>
          </div>
        )}
      </div>

      {/* 问题理解确认对话框 */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>创建思维导图</DialogTitle>
            <DialogDescription>确认问题并开始分析</DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>标题</Label>
              <Textarea
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="思维导图标题（可选）"
                className="min-h-[60px]"
              />
            </div>
            <div className="space-y-2">
              <Label>问题描述</Label>
              <Textarea
                value={problem}
                onChange={(e) => setProblem(e.target.value)}
                className="min-h-[80px]"
              />
            </div>

            <div className="space-y-2">
              <Label>目标（可选）</Label>
              <Textarea
                value={goal}
                onChange={(e) => setGoal(e.target.value)}
                placeholder="您希望达到什么目标？"
                className="min-h-[60px]"
              />
            </div>
          </div>

          <DialogFooter className="flex-row justify-end space-x-2">
            <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={handleCreateMap} disabled={isLoading}>
              {isLoading && <Loader className="h-4 w-4 mr-2 animate-spin" />}
              创建思维导图
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
