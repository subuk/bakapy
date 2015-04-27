!function(namespace) {
  'use strict';

  function Gantt(elem, params) {
    this.$element = jQuery(elem);
    this.params = params || {};

    this.cssHiddenChartLine = this.params.cssHiddenChartLine || 'JS-Gantt-ChartLine-hidden';
    this.cssHiddenChartGuide = this.params.cssHiddenChartGuide || 'JS-Gantt-ChartGuide-hidden';
    this.cssReadyElement = this.params.cssReadyElement || 'JS-Gantt-ready';
    this.timePerPixel = this.params.timePerPixel || null;
    this.timeStep = this.params.timeStep || 24 * 60 * 60 * 1000;

    this.__construct();
  }

  Gantt.prototype.__construct = function __construct() {
    if (typeof Modal === 'undefined') {
      return false;
    }

    this.$viewport = this.$element.find('.JS-Gantt-Viewport');
    this.$chart = this.$viewport.find('.JS-Gantt-Chart');
    this.$chartLineBlank = this.$chart.find('.JS-Gantt-ChartLine').eq(0);
    this.$chartFieldBlank = this.$chartLineBlank.find('.JS-Gantt-ChartField').eq(0);
    this.$chartPartBlank = this.$chartFieldBlank.find('.JS-Gantt-ChartPart').eq(0);
    this.$chartGuideBlank = this.$element.find('.JS-Gantt-ChartGuide').eq(0);
    this.$sidebar = this.$element.find('.JS-Gantt-Sidebar');
    this.$sidebarLineBlank = this.$sidebar.find('.JS-Gantt-SidebarLine');
    this.$modal = jQuery('.JS-Gannt-Modal');

    //modal elements
    this.$container = this.$modal.find('.JS-Modal-Container');
    this.$viewportM = this.$modal.find('.JS-Gantt-m-Viewport');
    this.$chartM = this.$viewportM.find('.JS-Gantt-m-Chart');
    this.$chartLineBlankM = this.$chartM.find('.JS-Gantt-m-ChartLine').eq(0);
    this.$sidebarM = this.$modal.find('.JS-Gantt-m-Sidebar');
    this.$sidebarLineBlankM = this.$sidebarM.find('.JS-Gantt-m-SidebarLine');
    this.$chartGuideBlankM = this.$modal.find('.JS-Gantt-m-ChartGuide').eq(0);

    this.data = {};
    this.startTime = new Date().getTime();
    this.endTime = 0;
    this.pixelDensity = 1000 * 60;
    this.chartLineHeight = this.$chartLineBlank.height();

    this._init();
  };

  Gantt.prototype._init = function _init() {
    var _this = this;

    this.$chart.on('click.JS-Gantt', '.JS-Gantt-ChartField', function() { _this._showDetails.apply(_this, [jQuery(this)]); });

    this._start();
  } ;

  Gantt.prototype._start = function _start() {
    this._getData();
    this._searchStartEndTime();

    if (!this.timePerPixel) {
      this._getTimePerPixel();
    } else {
      this._setChartWidth();
      this._initDraggable();
    }

    this._drawChart();
    this._drawTimeGrid(this.startTime, this.endTime, this.timePerPixel, this.timeStep, this.$chart, this.$chartGuideBlank, 'JS-Gantt-ChartGuide-hidden', 0);

    this._ready();
  };

  Gantt.prototype._ready = function _ready() {
    // dev
    // console.log(this.data);

    this.$element
      .addClass(this.cssReadyElement)
      .addClass('JS-Gantt-ready');
  };

  Gantt.prototype._getData = function _getData() {
    var _this = this;

    jQuery.ajax({
      url: '/backups.json',
      type: 'GET',
      async: false,
      success: function(response) {
        _this.data = response;
      },
      error: function() {
        _this.data = {};
      }
    });
  };

  Gantt.prototype._searchStartEndTime = function _searchStartEndTime() {
    var key,
        job,
        task,
        startTime,
        endTime,
        i,
        j;

    for (key in this.data) {
      job = this.data[key];

      for (i = 0, j = job.length; i < j; i++) {
        task = job[i];

        if (typeof task.StartTime !== 'undefined') {
          startTime = new Date(task.StartTime).getTime();

          if ( startTime < this.startTime) {
            this.startTime = startTime;
          }
        }

        if (typeof task.EndTime !== 'undefined') {
          endTime = new Date(task.EndTime).getTime();

          if ( endTime > this.endTime) {
            this.endTime = endTime;
          }
        }
      }
    }
  };

  Gantt.prototype._getTimePerPixel = function _getTimePerPixel() {
    this.timePerPixel = (this.endTime - this.startTime) / this.$chart.width();
  };

  Gantt.prototype._setChartWidth = function _setChartWidth() {
    var width = (this.endTime - this.startTime) / this.timePerPixel;
    this.$chart.css({width: Math.ceil(width) + 'px'});
  };

  Gantt.prototype._initDraggable = function _initDraggable() {
    this.$chart.draggable({
      axis: 'x',
      cursor: 'e-resize',
      scroll: true
    });
  };

  Gantt.prototype._drawChart = function _drawChart() {
    var key,
        job,
        task,
        taskPosition,
        taskCount,
        i,
        j;

    for (key in this.data) {
      job = this.data[key];

      // draw task
      for (i = 0, j = job.length; i < j; i++) {
        task = job[i];

        if (typeof task.StartTime !== 'undefined' && typeof task.EndTime !== 'undefined') {
          taskPosition = this._getTaskPosition(task.StartTime, task.EndTime, this.startTime, this.timePerPixel);
          this._drawTask(taskPosition, task.TaskId, task.JobName, this.$chart, this.$chartLineBlank, '.JS-Gantt-ChartField', 'JS-Gantt-ChartLine-hidden', this._niceFileSize(task.TotalSize));
        }
      }

      // draw job
      this._drawJob(key, j, this.$sidebar, this.$sidebarLineBlank, 'JS-Gantt-SidebarLine-hidden');
    }
  };

  Gantt.prototype._getTaskPosition = function _getTaskPosition(startTime, endTime, startTimeCommon, timePerPixel) {
    if (arguments.length !== 4) {
      return false;
    }

    var taskStart,
        taskEnd,
        startPosition,
        endPosition,
        taskWidth,
        taskDuration;

    taskStart = new Date(startTime).getTime() - startTimeCommon;
    taskEnd = new Date(endTime).getTime() - startTimeCommon;
    taskDuration = moment.utc(taskEnd - taskStart).format("HH:mm:ss.SSS");

    startPosition = Math.floor(taskStart / timePerPixel);
    endPosition = Math.ceil(taskEnd / timePerPixel);
    taskWidth = endPosition - startPosition;

    if (taskWidth < 1) {
      taskWidth = 1;
    }

    return {start: startPosition, width: taskWidth, duration: taskDuration};
  };

  Gantt.prototype._drawTask = function _drawTask(taskPosition, taskId, jobName, chart, chartLineBlank, chartFieldClass, chartLineHiddenClass, fileSize) {
    if (arguments.length < 8) {
      return false;
    }

    var $line = chartLineBlank.clone(),
        $field = $line.find(chartFieldClass);

    $line
      .removeClass(this.cssHiddenChartLine)
      .removeClass(chartLineHiddenClass);

    chart.append($line);

    $field
      .css({
        'left': taskPosition.start + 'px',
        'width': taskPosition.width + 'px'
      })
      .attr('data-taskId', taskId)
      .attr('data-jobName', jobName)
      .attr('title', taskPosition.duration + '(' + fileSize + ')');
  };

  Gantt.prototype._drawJob = function _drawJob(key, taskCount, sidebar, sidebarLineBlank, chartLineHiddenClass) {
    if (arguments.length !== 5) {
      return false;
    }

    var $line = sidebarLineBlank.clone();

    $line
      .removeClass(this.cssHiddenChartLine)
      .removeClass(chartLineHiddenClass);

    $line.text(key);
    $line.attr('title', key);

    sidebar.append($line);

    $line.css({
      'height': this.chartLineHeight * taskCount + 'px'
    });
  };

  Gantt.prototype._drawTimeGrid = function _drawTimeGrid(startTime, endTime, timePerPixel, timeStep, chart, chartGuideBlank, chartGuideHiddenClass, modalFl) {
    var $guide,
        guidePosition,
        date,
        i;

    for (i = startTime; i < endTime; i += timeStep) {
      $guide = chartGuideBlank.clone();
      guidePosition = (i - startTime) / timePerPixel;

      if (!modalFl) {
        date = new Date(i);

        $guide.attr('data-label', date.getDate() + '/' + date.getMonth() + '/' + date.getFullYear()); 
      } else {
        $guide.attr('data-label', moment.utc(i).format("HH:mm:ss.SSS"));
      }

      $guide
        .removeClass(this.cssHiddenChartGuide)
        .removeClass(chartGuideHiddenClass);

      chart.append($guide);

      $guide.css({'left': Math.ceil(guidePosition) + 'px'});
    }
  };

  Gantt.prototype._showDetails = function _showDetails($source) {
    if (!arguments.length) {
      return false;
    }

    var taskId = $source.attr('data-taskId'),
        jobName = $source.attr('data-jobName'),
        info,
        i,
        j,
        content = '';

    if (taskId) {
      // search info
      for (i = 0, j = this.data[jobName].length; i < j; i++) {
        if (this.data[jobName][i].TaskId === taskId) {
          info = this.data[jobName][i];
        } else {
          continue;
        }
      }

      // append to popup
      this.$modal.trigger('modal:open');

      // show popup
      this._formingPopupGrid(info.Files);
    }
  };

  Gantt.prototype._formingPopupGrid = function _formingPopupGrid(files) {
    if (!arguments.length) {
      return false;
    }

    var i,
        startTime = moment(files[0].StartTime).valueOf(),
        endTime = moment(files[0].EndTime).valueOf(),
        content = '',
        timePerPixel,
        filePosition,
        file,
        timeStep;

    //clear chart and sidebar and guide
    $('.JS-Gantt-m-ChartLine:not(.JS-Gantt-m-ChartLine-hidden)').remove();
    $('.JS-Gantt-m-SidebarLine:not(.JS-Gantt-m-SidebarLine-hidden)').remove();
    $('.JS-Gantt-m-ChartGuide:not(.JS-Gantt-m-ChartGuide-hidden)').remove();

    //search start/end time
    for (i = 0; i < files.length; i++) {
      file = files[i];

      if (typeof file.StartTime !== 'undefined' && typeof file.EndTime !== 'undefined') {
        if (moment(file.StartTime).valueOf() < startTime) startTime = moment(file.StartTime).valueOf();
        if (moment(file.EndTime).valueOf() > endTime) endTime = moment(file.EndTime).valueOf(); 
      }
    }

    //scroll fix
    endTime = endTime + (endTime-startTime)/20;

    timePerPixel = (endTime - startTime) / this.$viewportM.width();

    // draw files
    for (i = 0; i < files.length; i++) {
      file = files[i];

      this._drawJob(file.Name, 1, this.$sidebarM, this.$sidebarLineBlankM, 'JS-Gantt-m-SidebarLine-hidden');

      if (typeof file.StartTime !== 'undefined' && typeof file.EndTime !== 'undefined') {
        filePosition = this._getTaskPosition(file.StartTime, file.EndTime, startTime, timePerPixel);
        this._drawTask(filePosition, 0, file.Name, this.$chartM, this.$chartLineBlankM, '.JS-Gantt-m-ChartField', 'JS-Gantt-m-ChartLine-hidden', this._niceFileSize(file.Size));
      }
    }

    timeStep = (endTime - startTime) / 3; //3 time point
    this._drawTimeGrid(startTime, endTime, timePerPixel, timeStep, this.$chartM, this.$chartGuideBlankM, 'JS-Gantt-m-ChartGuide-hidden', 1);
  };

  Gantt.prototype._niceFileSize = function _niceFileSize(size) {

    function zeroRemove(data)
    {
      return data.replace(/\.0{2}/, "").replace(/(\.[\d]{1})0{1,}/,"$1");
    }

    if(size <= 1024)
    {
      return size + " " + "B";
    }
    else
    {
      size = (size / 1024);
      if(size <= 1024)
      {
        return zeroRemove(size.toFixed(2)) + " " + "KB";
      }
      else
      {
        size = (size / 1024);
        if(size <= 1024)
        {
          return zeroRemove(size.toFixed(2)) + " " + "MB";
        }
        else
        {
          size = (size / 1024);
          if(size <= 1024)
          {
            return zeroRemove(size.toFixed(2)) + " " + "GB";
          }
          else
          {
            size = (size / 1024);
            return zeroRemove(size.toFixed(2)) + " " + "TB";
          }
        }
      }
    }
  },

  namespace.Gantt = Gantt;
}(this);
